package cluster

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/openshift/osdctl/cmd/common"
	"github.com/openshift/osdctl/internal/utils/globalflags"
	"github.com/openshift/osdctl/pkg/controller"
	"github.com/openshift/osdctl/pkg/utils"
)

var hdr = color.New(color.FgBlue, color.Bold).SprintFunc()

// nolint:gosec
const pullSecStatusUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}

Required Flags:
  -C, --cluster-id string   Any cluster owned by the account (used to resolve the owner)
      --reason string        Elevation reason for cluster connections

Optional Flags:
      --check strings        Validate specific clusters' pull secrets (repeatable, comma-separated, or "all")
`

type pullSecretSnapshotOptions struct {
	clusterID     string
	reason        string
	checkClusters []string
	logger        *logrus.Logger

	genericclioptions.IOStreams
	GlobalOptions *globalflags.GlobalOptions
}

func newCmdPullSecretAudit(streams genericclioptions.IOStreams, globalOpts *globalflags.GlobalOptions) *cobra.Command {
	ops := &pullSecretSnapshotOptions{
		IOStreams:      streams,
		GlobalOptions: globalOpts,
		logger:        newSnapshotLogger(),
	}
	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Show pull secret status for all clusters owned by an account",
		Long: `Show pull secret status for all clusters sharing the same OCM account.

Given any cluster ID, resolves the owner account and lists all clusters
owned by that account. Compares cluster creation dates against the account's
registry credential update timestamps to flag clusters that may have stale
pull secrets.

Use --check to drill into specific clusters and compare their pull secrets
against the current OCM access token auth entries.`,
		Example: `  # Overview of all clusters for the account that owns this cluster
  osdctl cluster pull-secret audit -C 1kfmyclusterid --reason "OHSS-1234"

  # Validate a specific cluster's pull secret
  osdctl cluster pull-secret audit -C 1kfmyclusterid --reason "OHSS-1234" --check 2abcothercluster

  # Validate multiple clusters
  osdctl cluster pull-secret audit -C 1kfmyclusterid --reason "OHSS-1234" --check clusterA --check clusterB

  # Validate all clusters for this account
  osdctl cluster pull-secret audit -C 1kfmyclusterid --reason "OHSS-1234" --check all`,
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ops.run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&ops.clusterID, "cluster-id", "C", "", "Any cluster owned by the account (used to resolve the owner)")
	cmd.Flags().StringVar(&ops.reason, "reason", "", "Elevation reason for cluster connections")
	cmd.Flags().StringArrayVar(&ops.checkClusters, "check", nil, "Validate specific clusters' pull secrets (repeatable, or \"all\")")

	_ = cmd.MarkFlagRequired("cluster-id")
	_ = cmd.MarkFlagRequired("reason")

	cmd.SetUsageTemplate(pullSecStatusUsageTemplate)

	return cmd
}

func newSnapshotLogger() *logrus.Logger {
	l := logrus.New()
	l.SetOutput(os.Stderr)
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
		ForceColors:     true,
	})
	l.SetLevel(logrus.InfoLevel)
	return l
}

func (o *pullSecretSnapshotOptions) run(ctx context.Context) error {
	out := o.Out
	logger := o.logger

	log.SetLogger(zap.New(zap.WriteTo(o.ErrOut), zap.Level(zapcore.WarnLevel)))

	logger.Info("Creating OCM connection")
	ocm, err := utils.CreateConnection()
	if err != nil {
		return fmt.Errorf("failed to create OCM client: %w", err)
	}
	defer func() {
		if closeErr := ocm.Close(); closeErr != nil {
			logger.Warnf("Cannot close the OCM connection: %v", closeErr)
		}
	}()

	cluster, err := utils.GetClusterAnyStatus(ocm, o.clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}
	logger.Infof("Cluster resolved: %s (%s)", cluster.Name(), cluster.ID())

	subscription, err := utils.GetSubscription(ocm, cluster.ID())
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	ownerAccount, err := utils.GetAccount(ocm, subscription.Creator().ID())
	if err != nil {
		return fmt.Errorf("failed to get owner account: %w", err)
	}
	ownerAccountID := ownerAccount.ID()
	ownerUsername := ownerAccount.Username()
	logger.Infof("Owner resolved: %s (account: %s)", ownerUsername, ownerAccountID)

	logger.Info("Fetching registry credentials from OCM")
	latestCredUpdate, err := controller.GetLatestCredentialUpdate(ocm, ownerAccountID)
	if err != nil {
		logger.Warnf("Could not fetch registry credentials: %v", err)
	}

	logger.Info("Querying clusters for this account")
	clusters, err := controller.ListOwnerSubscriptions(ocm, ownerAccountID)
	if err != nil {
		return fmt.Errorf("failed to list subscriptions: %w", err)
	}

	// Resolve and validate --check targets
	checkIDs := resolveCheckTargets(o.checkClusters, clusters)
	clusterIDSet := make(map[string]bool, len(clusters))
	for _, c := range clusters {
		clusterIDSet[c.ID] = true
	}
	for _, id := range checkIDs {
		if !clusterIDSet[id] {
			return fmt.Errorf("cluster %s is not owned by account %s (%s)", id, ownerUsername, ownerAccountID)
		}
	}

	// --- Fetch OCM data for checks (graceful degradation) ---

	type checkResult struct {
		accessTokenResult *controller.PullSecretVerifyResult
		regCredResult     *controller.PullSecretVerifyResult
		err               error
	}
	checkResults := make(map[string]*checkResult)

	var auths map[string]*amv1.AccessTokenAuth
	hasAccessToken := false
	hasRegCreds := false

	if len(checkIDs) > 0 {
		// Try access token first — may fail without region-lead permissions
		logger.Infof("Fetching access token from OCM for owner '%s'", ownerUsername)
		_, auths, err = controller.FetchOwnerAccessToken(ocm, ownerUsername, logger)
		if err != nil {
			logger.Warnf("Could not fetch access token: %v", err)
			fmt.Fprintf(out, "\n%s Could not fetch OCM access token (may require region-lead permissions).\n", colorWarn("[WARN]"))
			fmt.Fprint(out, "Continue with registry credentials only? ")
			if !utils.ConfirmPrompt() {
				checkIDs = nil
			}
		} else {
			hasAccessToken = true
			logger.Infof("Retrieved %d auth entries from OCM access token", len(auths))
		}

		// Try registry credentials
		if len(checkIDs) > 0 {
			logger.Info("Fetching registry credentials from OCM")
			testCreds, regErr := utils.GetRegistryCredentials(ocm, ownerAccountID)
			if regErr != nil || len(testCreds) == 0 {
				logger.Warnf("Could not fetch registry credentials: %v", regErr)
			} else {
				hasRegCreds = true
				logger.Infof("Retrieved %d registry credentials from OCM", len(testCreds))
			}
		}

		// If both failed, skip checks
		if !hasAccessToken && !hasRegCreds {
			logger.Warn("Neither access token nor registry credentials available — skipping checks")
			fmt.Fprintf(out, "%s Cannot compare cluster pull secrets without OCM data. Skipping --check.\n", colorWarn("[WARN]"))
			checkIDs = nil
		}
	}

	// --- Connect to clusters and collect results ---

	if len(checkIDs) > 0 {
		elevationReasons := []string{
			o.reason,
			"Checking pull secret status using osdctl pull-secret snapshot",
		}

		for _, clusterID := range checkIDs {
			cr := &checkResult{}
			logger.Infof("Connecting to cluster %s", clusterID)
			_, _, clientset, connErr := common.GetKubeConfigAndClient(clusterID, elevationReasons...)
			if connErr != nil {
				cr.err = fmt.Errorf("failed to connect: %v", connErr)
				checkResults[clusterID] = cr
				continue
			}

			if hasAccessToken {
				result, verifyErr := controller.CompareAccessTokenAuthsToCluster(ctx, clientset, auths, nil)
				if verifyErr != nil {
					logger.Warnf("Access token verification failed for %s: %v", clusterID, verifyErr)
				} else {
					cr.accessTokenResult = result
				}
			}

			if hasRegCreds {
				result, verifyErr := controller.CompareRegistryCredentialAuthsToCluster(ctx, ocm, clientset, ownerAccountID, ownerAccount.Email(), nil)
				if verifyErr != nil {
					logger.Warnf("Registry credential verification failed for %s: %v", clusterID, verifyErr)
				} else {
					cr.regCredResult = result
				}
			}

			checkResults[clusterID] = cr
		}
	}

	// --- Render ---

	fmt.Fprintf(out, "\n============================================================\n")
	fmt.Fprintf(out, " Owner:    %s (account: %s)\n", ownerUsername, ownerAccountID)
	fmt.Fprintf(out, " Email:    %s\n", ownerAccount.Email())
	if !latestCredUpdate.IsZero() {
		fmt.Fprintf(out, " Registry credentials last updated: %s\n", latestCredUpdate.Format("2006-01-02 15:04:05 UTC"))
	}
	fmt.Fprintf(out, "============================================================\n\n")

	staleCount := 0
	for i, c := range clusters {
		if i > 0 {
			fmt.Fprintln(out)
			fmt.Fprintln(out, "============================================================")
			fmt.Fprintln(out)
		}
		renderClusterBanner(out, c)

		cr, hasCheck := checkResults[c.ID]

		if hasCheck && cr.err != nil {
			renderPSStatus(out, c, latestCredUpdate, &staleCount)
			fmt.Fprintf(out, "  %s PS CHECK: %v\n", colorFail("[FAIL]"), cr.err)
		} else if hasCheck {
			if cr.accessTokenResult != nil {
				renderCheckTable(out, cr.accessTokenResult, c.ID, "ACCESS TOKEN AUTHS")
			}
			if cr.regCredResult != nil {
				renderCheckTable(out, cr.regCredResult, c.ID, "REGISTRY CREDENTIAL AUTHS")
			}
		} else {
			renderPSStatus(out, c, latestCredUpdate, &staleCount)
		}
		fmt.Fprintln(out)
	}

	fmt.Fprintf(out, "%d cluster(s) found", len(clusters))
	if staleCount > 0 {
		fmt.Fprintf(out, ", %s %d potentially stale", colorWarn("[WARN]"), staleCount)
	}
	if len(checkIDs) > 0 {
		fmt.Fprintf(out, ", %d checked", len(checkIDs))
	}
	fmt.Fprintln(out, ".")

	if len(o.checkClusters) == 0 {
		fmt.Fprintf(out, "Use --check <cluster-id> or --check all to validate pull secrets against OCM.\n")
	}

	return nil
}

func renderClusterBanner(out io.Writer, c controller.ClusterSummary) {
	label := color.New(color.FgBlue, color.Bold).SprintFunc()
	fmt.Fprintf(out, "%s  %s (%s)\n", label("Cluster:"), c.Name, c.ID)
	fmt.Fprintf(out, "%s  %s    %s  %s\n",
		label("Created:"), c.CreatedAt.Format("2006-01-02 15:04"),
		label("Status:"), c.Status)
}

func renderPSStatus(out io.Writer, c controller.ClusterSummary, latestCredUpdate time.Time, staleCount *int) {
	psLabel := color.New(color.FgCyan, color.Bold).SprintFunc()
	psDetail := color.New(color.FgCyan).SprintFunc()

	if latestCredUpdate.IsZero() {
		fmt.Fprintf(out, "  %s %s\n", psLabel("PS STATUS:"), psDetail("unknown, no registry credential timestamps available"))
	} else if c.CreatedAt.Before(latestCredUpdate) {
		fmt.Fprintf(out, "  %s %s %s\n", colorWarn("[WARN]"), psLabel("PS STATUS:"), psDetail("may be stale, created before registry credentials were last updated"))
		*staleCount++
	} else {
		fmt.Fprintf(out, "  %s %s\n", psLabel("PS STATUS:"), psDetail("unknown, but created after registry credentials were last updated"))
	}
}

func renderCheckTable(out io.Writer, result *controller.PullSecretVerifyResult, clusterID string, sourceLabel string) {
	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{sourceLabel, "TOKEN", "EMAIL", "STATUS"})
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.SetColumnSeparator("  ")
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlueColor},
	)

	mismatchStatus := color.New(color.FgYellow, color.Bold).SprintFunc()
	for _, ar := range result.AuthResults {
		status := colorOK("[OK]")
		tokenStr := "match"
		emailStr := "match"
		if !ar.OK {
			status = mismatchStatus("[!]")
			if ar.Detail == "not found in cluster secret" {
				tokenStr = "missing"
				emailStr = "missing"
			} else {
				if !ar.TokenMatch {
					tokenStr = "MISMATCH"
				}
				if !ar.EmailMatch {
					emailStr = "MISMATCH"
				}
			}
		}
		table.Append([]string{ar.Registry, tokenStr, emailStr, status})
	}
	table.Render()

	if result.Matched < result.Total {
		fmt.Fprintf(out, "  %s Verified %d/%d — consider 'osdctl cluster pull-secret update -C %s'\n",
			colorWarn("[WARN]"), result.Matched, result.Total, clusterID)
	}
	if len(result.MissingRequired) > 0 {
		fmt.Fprintf(out, "  %s missing required registries: %s\n",
			colorWarn("[WARN]"), strings.Join(result.MissingRequired, ", "))
	}
}

func resolveCheckTargets(checkArgs []string, clusters []controller.ClusterSummary) []string {
	var ids []string
	for _, arg := range checkArgs {
		if strings.EqualFold(arg, "all") {
			for _, c := range clusters {
				ids = append(ids, c.ID)
			}
			return ids
		}
		for _, part := range strings.Split(arg, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				ids = append(ids, part)
			}
		}
	}
	return ids
}

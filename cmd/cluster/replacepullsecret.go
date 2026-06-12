package cluster

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/fatih/color"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/openshift/osdctl/cmd/common"
	"github.com/openshift/osdctl/cmd/servicelog"
	"github.com/openshift/osdctl/internal/utils/globalflags"
	"github.com/openshift/osdctl/pkg/controller"
	"github.com/openshift/osdctl/pkg/utils"
)

var (
	reasonPattern = regexp.MustCompile(`(?i)(OHSS|PD|SREP|OSD|SDE|ROSAENG)-\d+`)

	colorOK     = color.New(color.FgGreen).SprintFunc()
	colorFail   = color.New(color.FgRed).SprintFunc()
	colorWarn   = color.New(color.FgYellow).SprintFunc()
	colorDryRun = color.New(color.FgCyan).SprintFunc()
)

// nolint:gosec
const replacePullSecUsageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}

Required Flags:
  -C, --cluster-id string   The Internal/External Cluster ID or Cluster Name
      --reason string        The reason for this command (usually an OHSS or PD ticket)

Optional Flags:
  -d, --dry-run              Dry-run - show what would change but do not apply
      --hive-ocm-url string  OCM environment for Hive operations (aliases: production, staging, integration)
`

type replacePullSecretOptions struct {
	clusterID  string
	reason     string
	dryrun     bool
	hiveOcmUrl string
	logger     *logrus.Logger

	genericclioptions.IOStreams
	GlobalOptions *globalflags.GlobalOptions
}

// dryRunChecker tracks dry-run status and provides formatted output methods.
type dryRunChecker struct {
	out   io.Writer
	allOK bool
}

func (d *dryRunChecker) would(format string, args ...any) {
	fmt.Fprintf(d.out, "%s %s %s\n", colorDryRun("[Dry Run]"), colorDryRun("Would:"), colorDryRun(fmt.Sprintf(format, args...)))
}

func (d *dryRunChecker) report(ok bool, format string, args ...any) {
	status := colorOK("[OK]")
	if !ok {
		status = colorFail("[FAIL]")
		d.allOK = false
	}
	fmt.Fprintf(d.out, "%s %s %s\n", colorDryRun("[Dry Run]"), status, fmt.Sprintf(format, args...))
}

func (d *dryRunChecker) info(format string, args ...any) {
	fmt.Fprintf(d.out, "%s %s\n", colorDryRun("[Dry Run]"), fmt.Sprintf(format, args...))
}

func (d *dryRunChecker) section(title string, lines ...string) {
	detail := color.New(color.FgWhite).SprintFunc()
	fmt.Fprintf(d.out, "\n%s\n", colorDryRun("============================================================"))
	fmt.Fprintf(d.out, "%s %s\n", colorDryRun("[Dry Run]"), colorDryRun(title))
	for _, line := range lines {
		fmt.Fprintf(d.out, "  %s\n", detail(line))
	}
	fmt.Fprintf(d.out, "%s\n", colorDryRun("============================================================"))
}

func newReplacePullSecretLogger() *logrus.Logger {
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

func newCmdReplacePullSecretDeprecated(streams genericclioptions.IOStreams, globalOpts *globalflags.GlobalOptions) *cobra.Command {
	cmd := newCmdPullSecretUpdate(streams, globalOpts)
	cmd.Use = "replace-pull-secret"
	cmd.Deprecated = "use 'osdctl cluster pull-secret update' instead"
	return cmd
}

func newCmdPullSecretUpdate(streams genericclioptions.IOStreams, globalOpts *globalflags.GlobalOptions) *cobra.Command {
	ops := &replacePullSecretOptions{
		IOStreams:      streams,
		GlobalOptions: globalOpts,
		logger:        newReplacePullSecretLogger(),
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Refresh a cluster's pull secret from the cluster owner's OCM account",
		Long: `Refresh a cluster's pull secret from the cluster owner's OCM account.

This updates the pull secret on a ROSA HCP or Classic cluster without performing
an ownership transfer. The pull secret is rebuilt from the latest credentials
in the cluster owner's OCM account.

See documentation prior to executing:
https://github.com/openshift/ops-sop/blob/master/hypershift/knowledge_base/howto/replace-pull-secret.md
https://github.com/openshift/ops-sop/blob/master/v4/howto/transfer_cluster_ownership.md`,
		Example: `  # Replace pull secret on a cluster
  osdctl cluster pull-secret update --cluster-id 1kfmyclusterid --reason "OHSS-1234"

  # Dry-run to preview without making changes
  osdctl cluster pull-secret update --cluster-id 1kfmyclusterid --reason "OHSS-1234" --dry-run`,
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ops.validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ops.run(cmd.Context())
		},
	}

	cmd.Flags().StringVarP(&ops.clusterID, "cluster-id", "C", "", "The Internal/External Cluster ID or Cluster Name")
	cmd.Flags().StringVar(&ops.reason, "reason", "", "The reason for this command (usually an OHSS or PD ticket)")
	cmd.Flags().BoolVarP(&ops.dryrun, "dry-run", "d", false, "Dry-run - show what would change but do not apply")
	cmd.Flags().StringVar(&ops.hiveOcmUrl, "hive-ocm-url", "", "OCM environment for Hive operations (aliases: production, staging, integration)")

	_ = cmd.MarkFlagRequired("cluster-id")
	_ = cmd.MarkFlagRequired("reason")

	cmd.SetUsageTemplate(replacePullSecUsageTemplate)

	return cmd
}

func (o *replacePullSecretOptions) validate() error {
	if !reasonPattern.MatchString(o.reason) {
		o.logger.Warnf("--reason %q does not appear to contain a ticket ID (e.g. OHSS-1234)", o.reason)
		fmt.Fprint(o.Out, "Continue without a valid ticket reference? ")
		if !utils.ConfirmPrompt() {
			return fmt.Errorf("operation aborted — provide a valid --reason")
		}
	}
	if o.hiveOcmUrl != "" {
		resolved, err := utils.ValidateAndResolveOcmUrl(o.hiveOcmUrl)
		if err != nil {
			return fmt.Errorf("invalid --hive-ocm-url: %w", err)
		}
		o.hiveOcmUrl = resolved
	}
	return nil
}

func (o *replacePullSecretOptions) run(ctx context.Context) error {
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

	// --- Phase 1: Identify cluster and confirm with user ---

	cluster, err := utils.GetClusterAnyStatus(ocm, o.clusterID)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}
	o.clusterID = cluster.ID()
	logger.Infof("Cluster resolved: %s (%s)", cluster.Name(), o.clusterID)

	isHCP, err := utils.IsHostedCluster(o.clusterID)
	if err != nil {
		return fmt.Errorf("failed to check if cluster is HCP: %w", err)
	}

	clusterType := "OSD/ROSA Classic"
	if isHCP {
		clusterType = "HCP"
	}

	subscription, err := utils.GetSubscription(ocm, o.clusterID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	ownerAccount, err := utils.GetAccount(ocm, subscription.Creator().ID())
	if err != nil {
		return fmt.Errorf("failed to get owner account from subscription: %w", err)
	}
	ownerUsername := ownerAccount.Username()
	ownerAccountID := ownerAccount.ID()

	siblingCount := controller.CountOwnerClusters(ocm, ownerAccountID, logger)

	fmt.Fprintf(out, "\n============================================================\n")
	fmt.Fprintf(out, " Cluster:  %s (%s)\n", cluster.Name(), o.clusterID)
	fmt.Fprintf(out, " Type:     %s\n", clusterType)
	fmt.Fprintf(out, " Owner:    %s (account: %s)\n", ownerUsername, ownerAccountID)
	fmt.Fprintf(out, " Reason:   %s\n", o.reason)
	if o.dryrun {
		fmt.Fprintf(out, " Mode:     %s\n", colorDryRun("DRY-RUN (no changes will be made)"))
	}
	if siblingCount > 1 {
		fmt.Fprintf(out, "\n %s This account owns %d clusters sharing the same access token.\n", colorWarn("[NOTE]"), siblingCount)
		fmt.Fprintf(out, "          This command only updates the pull secret on the cluster above.\n")
		fmt.Fprintf(out, "          Use 'osdctl cluster pull-secret audit -C %s' to review all clusters for this account.\n", o.clusterID)
	}
	fmt.Fprintf(out, "============================================================\n")
	fmt.Fprint(out, "Is this the correct cluster? ")
	if !utils.ConfirmPrompt() {
		return fmt.Errorf("operation aborted by user")
	}

	// --- Phase 2: Resolve infrastructure clusters ---

	var hiveOCM *sdk.Connection
	if o.hiveOcmUrl != "" {
		logger.Infof("Creating separate OCM connection for Hive operations: %s", o.hiveOcmUrl)
		hiveOCM, err = utils.CreateConnectionWithUrl(o.hiveOcmUrl)
		if err != nil {
			if !o.dryrun {
				return fmt.Errorf("failed to create hive OCM connection with URL '%s': %w", o.hiveOcmUrl, err)
			}
			logger.Warnf("Failed to create hive OCM connection: %v", err)
		} else {
			defer hiveOCM.Close()
		}
	}

	var mgmtCluster *cmv1.Cluster
	var masterCluster *cmv1.Cluster
	var infraResolved bool

	if isHCP {
		logger.Info("Resolving HCP Management and Service clusters")
		mgmtCluster, err = utils.GetManagementCluster(o.clusterID)
		if err != nil {
			if !o.dryrun {
				return fmt.Errorf("failed to get management cluster: %w", err)
			}
			logger.Warnf("Failed to get management cluster: %v", err)
		}
		svcCluster, err := utils.GetServiceCluster(o.clusterID)
		if err != nil {
			if !o.dryrun {
				return fmt.Errorf("failed to get service cluster: %w", err)
			}
			logger.Warnf("Failed to get service cluster: %v", err)
		}
		if svcCluster != nil {
			masterCluster = svcCluster
			infraResolved = true
			fmt.Fprintf(out, "  Management cluster: %s\n", mgmtCluster.Name())
			fmt.Fprintf(out, "  Service cluster:    %s\n", svcCluster.Name())
		}
	} else {
		logger.Info("Resolving Hive cluster")
		var hiveCluster *cmv1.Cluster
		if hiveOCM != nil {
			hiveCluster, err = utils.GetHiveClusterWithConn(o.clusterID, ocm, hiveOCM)
		} else {
			hiveCluster, err = utils.GetHiveCluster(o.clusterID)
		}
		if err != nil {
			if !o.dryrun {
				return fmt.Errorf("failed to get hive cluster: %w", err)
			}
			logger.Warnf("Failed to get hive cluster: %v", err)
		} else {
			masterCluster = hiveCluster
			infraResolved = true
			fmt.Fprintf(out, "  Hive cluster: %s\n", hiveCluster.Name())
		}
	}

	// --- Phase 3: Fetch pull secret from OCM ---

	logger.Infof("Fetching pull secret from OCM for owner '%s'", ownerUsername)
	var pullSecret []byte
	var auths map[string]*amv1.AccessTokenAuth
	pullSecret, auths, err = controller.FetchOwnerPullSecret(ocm, ownerUsername, logger)
	if err != nil {
		if !o.dryrun {
			return err
		}
		logger.Warnf("Failed to fetch OCM access token: %v", err)
	} else {
		logger.Infof("Retrieved %d auth entries from OCM access token", len(auths))

		missingFromOCM := controller.ValidateRequiredAuths(auths)
		if len(missingFromOCM) > 0 {
			logger.Warn("OCM access token is missing required auth entries")
			for _, m := range missingFromOCM {
				fmt.Fprintf(out, "  %s missing: %s\n", colorWarn("[WARN]"), m)
			}
			fmt.Fprintf(out, "This may indicate an issue with the cluster owner's OCM account.\n")
			if !o.dryrun {
				fmt.Fprint(out, "Continue with incomplete pull secret? ")
				if !utils.ConfirmPrompt() {
					return fmt.Errorf("aborted — OCM access token missing required registries: %v", missingFromOCM)
				}
			}
		}
	}

	// --- Phase 4: Connect to clusters ---

	elevationReasons := []string{
		o.reason,
		"Replacing pull secret using osdctl pull-secret update",
	}

	var masterKubeCli client.Client
	var masterKubeClientSet *kubernetes.Clientset
	var infraConnected bool

	if infraResolved {
		logger.Infof("Connecting to infrastructure cluster %s (%s)", masterCluster.Name(), masterCluster.ID())
		if hiveOCM != nil {
			masterKubeCli, _, masterKubeClientSet, err = common.GetKubeConfigAndClientWithConn(masterCluster.ID(), hiveOCM, elevationReasons...)
		} else {
			masterKubeCli, _, masterKubeClientSet, err = common.GetKubeConfigAndClient(masterCluster.ID(), elevationReasons...)
		}
		if err != nil {
			if !o.dryrun {
				return fmt.Errorf("failed to get kube client for infrastructure cluster %s: %w", masterCluster.ID(), err)
			}
			logger.Warnf("Failed to connect to infrastructure cluster: %v", err)
		} else {
			infraConnected = true
		}
	}

	var targetClientSet *kubernetes.Clientset
	var targetConnected bool

	logger.Infof("Connecting to target cluster %s (%s)", cluster.Name(), o.clusterID)
	_, _, targetClientSet, err = common.GetKubeConfigAndClient(o.clusterID, elevationReasons...)
	if err != nil {
		if !o.dryrun {
			return fmt.Errorf("failed to get kube client for target cluster %s: %w", o.clusterID, err)
		}
		logger.Warnf("Failed to connect to target cluster: %v", err)
	} else {
		targetConnected = true
	}

	// --- Dry-run walkthrough ---

	if o.dryrun {
		mgmtName := ""
		if mgmtCluster != nil {
			mgmtName = mgmtCluster.Name()
		}
		infraName := "(unresolved)"
		if masterCluster != nil {
			infraName = masterCluster.Name()
		}
		if err := dryRunWalkthrough(ctx, masterKubeCli, masterKubeClientSet, targetClientSet,
			isHCP, o.clusterID, cluster.Name(), infraName, mgmtName,
			auths, out,
			infraResolved, infraConnected, targetConnected); err != nil {
			return err
		}
		fmt.Fprintf(out, "\n%s Pull secret replacement pre-check completed (no changes made)\n", colorDryRun("[Dry Run]"))
		return nil
	}

	preflight, err := controller.PreflightCheck(ctx, targetClientSet, isHCP, cluster.Name(), out)
	if err != nil {
		return err
	}
	_ = preflight // TODO: use preflight.SecretData for merge logic

	// --- Phase 5: Apply pull secret update ---

	fmt.Fprint(out, "\nProceed with pull secret replacement? ")
	if !utils.ConfirmPrompt() {
		return fmt.Errorf("operation aborted by user")
	}

	logger.Info("Applying pull secret update")
	if isHCP {
		err = updateManifestWork(ocm, masterKubeCli, o.clusterID, mgmtCluster.Name(), pullSecret)
		if err != nil {
			return fmt.Errorf("failed to update pull secret via ManifestWork: %w", err)
		}
	} else {
		err = updatePullSecret(ocm, masterKubeCli, masterKubeClientSet, o.clusterID, pullSecret)
		if err != nil {
			return fmt.Errorf("failed to update pull secret via Hive SyncSet: %w", err)
		}
	}

	if !isHCP {
		logger.Info("Rolling out pods openshift-monitoring/telemeter-client")
		if err := rolloutPods(targetClientSet, "openshift-monitoring", "app.kubernetes.io/name=telemeter-client"); err != nil {
			logger.Warnf("Failed to roll out pods openshift-monitoring/telemeter-client: %v", err)
		}
	}

	// --- Phase 6: Post-operation verification ---

	logger.Infof("Verifying secret openshift-config/pull-secret on %s", cluster.Name())
	result, err := controller.VerifyPullSecretAuths(ctx, targetClientSet, auths, out)
	if err != nil {
		return fmt.Errorf("post-operation verification failed: %w", err)
	}
	if len(result.Mismatches) > 0 {
		fmt.Fprintf(out, "  %s %d auth(s) differ from OCM: %v\n", colorFail("[FAIL]"), len(result.Mismatches), result.Mismatches)
		fmt.Fprint(out, "Continue despite mismatches? ")
		if !utils.ConfirmPrompt() {
			return fmt.Errorf("verification failed — %d auth entries did not match", len(result.Mismatches))
		}
	} else {
		fmt.Fprintf(out, "  %s All OCM auth entries verified on target cluster\n", colorOK("[OK]"))
	}

	if !isHCP {
		logger.Info("Rolling out pods openshift-ocm-agent-operator/ocm-agent")
		if err := rolloutPods(targetClientSet, "openshift-ocm-agent-operator", "app=ocm-agent"); err != nil {
			logger.Warnf("Failed to roll out pods openshift-ocm-agent-operator/ocm-agent: %v", err)
		}
	}

	// --- Phase 7: Service log ---

	logger.Info("Sending internal service log")
	postCmd := servicelog.PostCmdOptions{
		ClusterId: o.clusterID,
		TemplateParams: []string{
			fmt.Sprintf("MESSAGE=Pull secret replaced for cluster owner '%s'. Reason: %s", ownerUsername, o.reason),
		},
		InternalOnly: true,
	}
	if err := postCmd.Run(); err != nil {
		logger.Warnf("Failed to send internal service log: %v", err)
		fmt.Fprintf(out, "Please manually send: osdctl servicelog post -i %s -p MESSAGE=\"Pull secret replaced for cluster owner '%s'.\"\n", o.clusterID, ownerUsername)
	}

	fmt.Fprintf(out, "\n%s Pull secret replacement completed successfully\n", colorOK("[OK]"))
	return nil
}

// dryRunWalkthrough walks through every action the live run would perform,
// interleaving "Would:" statements with RBAC and resource existence checks.
func dryRunWalkthrough(ctx context.Context, masterKubeCli client.Client, infraClientSet *kubernetes.Clientset, targetClientSet *kubernetes.Clientset, isHCP bool, clusterID string, targetName string, infraName string, mgmtClusterName string, auths map[string]*amv1.AccessTokenAuth, out io.Writer, infraResolved bool, infraConnected bool, targetConnected bool) error {
	dr := &dryRunChecker{out: out, allOK: true}

	// --- Step 1: Summary of OCM data and connectivity ---

	authStatus := fmt.Sprintf("%d auth entries fetched", len(auths))
	if auths == nil {
		authStatus = colorFail("[FAIL]") + " could not fetch access token (may require region-lead permissions)"
		dr.allOK = false
	}

	infraStatus := colorOK("[OK]") + " connected with elevation"
	if !infraResolved {
		infraStatus = colorFail("[FAIL]") + " could not resolve infrastructure cluster"
		dr.allOK = false
	} else if !infraConnected {
		infraStatus = colorFail("[FAIL]") + " could not connect to infrastructure cluster"
		dr.allOK = false
	}

	targetStatus := colorOK("[OK]") + " connected with elevation"
	if !targetConnected {
		targetStatus = colorFail("[FAIL]") + " could not connect to target cluster"
		dr.allOK = false
	}

	dr.section("Step 1: OCM data and cluster connectivity",
		"Verifies OCM access token can be fetched for the cluster owner and that",
		"connections can be established to both the infrastructure and target clusters.",
		"",
		fmt.Sprintf("  Cluster:              %s (%s)", targetName, clusterID),
		fmt.Sprintf("  Infrastructure:       %s", infraName),
		fmt.Sprintf("  OCM access token:     %s", authStatus),
		fmt.Sprintf("  Infra connection:     %s", infraStatus),
		fmt.Sprintf("  Target connection:    %s", targetStatus))

	// --- Section 2: Infrastructure cluster pull secret update ---

	// --- Step 2: Infrastructure cluster pull secret update ---

	if isHCP {
		dr.section("Step 2: Update pull secret via ManifestWork (HCP)",
			"HCP clusters store the pull secret inside a ManifestWork on the service cluster.",
			fmt.Sprintf("The ManifestWork %s/%s will be updated with new auth data.", mgmtClusterName, clusterID),
			"The work agent then syncs the secret from service cluster → management cluster → hosted cluster.")

		if infraConnected {
			dr.would("get and update ManifestWork %s/%s on service cluster %s", mgmtClusterName, clusterID, infraName)
			dr.canI(ctx, infraClientSet, infraName, "get", "manifestworks", "work.open-cluster-management.io", mgmtClusterName)
			dr.canI(ctx, infraClientSet, infraName, "update", "manifestworks", "work.open-cluster-management.io", mgmtClusterName)
		} else {
			dr.would("get and update ManifestWork %s/%s on service cluster %s", mgmtClusterName, clusterID, infraName)
			fmt.Fprintf(out, "  %s cannot verify — infrastructure cluster not connected\n", colorFail("[SKIP]"))
		}
	} else {
		dr.section("Step 2: Update pull secret via Hive SyncSet (Classic)",
			"Classic clusters store the pull secret in a Hive namespace on the hive cluster.",
			"The old secret is deleted and recreated, then a SyncSet syncs it to the target cluster.",
			"After sync completes, the SyncSet is cleaned up.")

		if infraConnected && masterKubeCli != nil {
			dr.info("Resolving Hive namespace for cluster %s on %s...", clusterID, infraName)
			hiveInfo, hiveErr := controller.FindHiveNamespace(ctx, masterKubeCli, clusterID)
			if hiveErr != nil {
				fmt.Fprintf(out, "  %s could not find Hive namespace: %v\n", colorFail("[FAIL]"), hiveErr)
				dr.allOK = false
			} else {
				hiveNS := hiveInfo.Namespace
				dr.report(true, "found ClusterDeployment %s/%s on %s", hiveNS, hiveInfo.ClusterDeploymentName, infraName)

				_, secretErr := infraClientSet.CoreV1().Secrets(hiveNS).Get(ctx, "pull", metav1.GetOptions{})
				if secretErr != nil {
					dr.report(false, "secret %s/pull not found on %s — will be created", hiveNS, infraName)
				} else {
					dr.report(true, "secret %s/pull exists on %s", hiveNS, infraName)
				}

				dr.would("delete secret %s/pull on %s", hiveNS, infraName)
				dr.canI(ctx, infraClientSet, infraName, "get", "secrets", "", hiveNS)
				dr.canI(ctx, infraClientSet, infraName, "delete", "secrets", "", hiveNS)

				dr.would("create secret %s/pull on %s with updated pull secret data", hiveNS, infraName)
				dr.canI(ctx, infraClientSet, infraName, "create", "secrets", "", hiveNS)

				dr.would("create SyncSet %s/pull-secret-replacement to sync to %s", hiveNS, targetName)
				dr.canI(ctx, infraClientSet, infraName, "create", "syncsets", "hive.openshift.io", hiveNS)

				dr.would("poll ClusterSync %s/%s then delete SyncSet", hiveNS, hiveInfo.ClusterDeploymentName)
				dr.canI(ctx, infraClientSet, infraName, "get", "clustersync", "hiveinternal.openshift.io", hiveNS)
				dr.canI(ctx, infraClientSet, infraName, "delete", "syncsets", "hive.openshift.io", hiveNS)
			}
		} else {
			dr.would("resolve Hive namespace, delete/create secret, create SyncSet on %s", infraName)
			fmt.Fprintf(out, "  %s cannot verify — infrastructure cluster not connected\n", colorFail("[SKIP]"))
		}
	}

	// --- Step 3: Pod rollouts (Classic only) ---

	if !isHCP {
		dr.section("Step 3: Pod rollouts (Classic only)",
			"After the pull secret is synced, telemeter-client and ocm-agent pods are restarted",
			"so they pick up the new credentials. HCP clusters do not require pod rollouts.")

		if targetConnected {
			dr.would("roll out pods openshift-monitoring/telemeter-client on %s", targetName)
			dr.canI(ctx, targetClientSet, targetName, "delete", "pods", "", "openshift-monitoring")

			dr.would("roll out pods openshift-ocm-agent-operator/ocm-agent on %s", targetName)
			dr.canI(ctx, targetClientSet, targetName, "delete", "pods", "", "openshift-ocm-agent-operator")
		} else {
			dr.would("roll out telemeter-client and ocm-agent pods on %s", targetName)
			fmt.Fprintf(out, "  %s cannot verify — target cluster not connected\n", colorFail("[SKIP]"))
		}
	}

	// --- Step N: Verification ---

	stepNum := "3"
	if !isHCP {
		stepNum = "4"
	}
	dr.section(fmt.Sprintf("Step %s: Verify pull secret on target cluster", stepNum),
		"After update, the pull secret on the target cluster is compared against the OCM",
		"access token to verify all auth entries match (token + email per registry).",
		"Required registries are also checked to ensure the cluster can pull images.")

	if targetConnected {
		dr.canI(ctx, targetClientSet, targetName, "get", "secrets", "", "openshift-config")
		dr.canI(ctx, targetClientSet, targetName, "update", "secrets", "", "openshift-config")

		if auths != nil {
			dr.info("Checking current state of secret openshift-config/pull-secret on %s...", targetName)
			result, err := controller.VerifyPullSecretAuths(ctx, targetClientSet, auths, out)
			if err != nil {
				fmt.Fprintf(out, "  %s current pull secret state: %v\n", colorWarn("[WARN]"), err)
			} else if result.Matched == result.Total {
				fmt.Fprintf(out, "  %s Pull secret is already up to date — a live run would be a no-op\n", colorOK("[INFO]"))
			}
		} else {
			fmt.Fprintf(out, "  %s cannot compare — OCM access token not available\n", colorFail("[SKIP]"))
		}
	} else {
		fmt.Fprintf(out, "  %s cannot verify — target cluster not connected\n", colorFail("[SKIP]"))
	}

	// --- Step N+1: Service log ---

	stepNum2 := "4"
	if !isHCP {
		stepNum2 = "5"
	}
	dr.section(fmt.Sprintf("Step %s: Send internal service log", stepNum2),
		"An internal (non-customer-visible) service log is sent to record that the",
		"pull secret was updated, including the owner username and reason.")

	dr.would("send internal service log for %s", targetName)

	if !dr.allOK {
		fmt.Fprintf(out, "\n%s %s Some checks failed. Verify elevated permissions before running without --dry-run.\n", colorDryRun("[Dry Run]"), colorFail("[FAIL]"))
	} else {
		fmt.Fprintf(out, "\n%s %s All pre-flight checks passed. No changes were made.\n", colorDryRun("[Dry Run]"), colorOK("OK"))
	}

	return nil
}

// canI checks RBAC permission and reports it in dry-run format with system label.
func (d *dryRunChecker) canI(ctx context.Context, clientset *kubernetes.Clientset, systemLabel, verb, resource, group, namespace string) {
	allowed, err := checkCanI(ctx, clientset, verb, resource, group, namespace)
	nsLabel := namespace
	if nsLabel == "" {
		nsLabel = "(cluster-scoped)"
	}
	if err != nil {
		fmt.Fprintf(d.out, "%s %s %s: auth can-i %s %s in %s (%v)\n",
			colorDryRun("[Dry Run]"), colorWarn("[SKIP]"), systemLabel, verb, resource, nsLabel, err)
	} else {
		d.report(allowed, "%s: auth can-i %s %s in %s", systemLabel, verb, resource, nsLabel)
	}
}

// checkCanI performs a SelfSubjectAccessReview to verify RBAC permission.
func checkCanI(ctx context.Context, clientset *kubernetes.Clientset, verb, resource, group, namespace string) (bool, error) {
	review := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Verb:      verb,
				Resource:  resource,
				Group:     group,
				Namespace: namespace,
			},
		},
	}
	result, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, review, metav1.CreateOptions{})
	if err != nil {
		return false, err
	}
	return result.Status.Allowed, nil
}

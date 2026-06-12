package cluster

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/fatih/color"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
      --force                Proceed despite dry-run failures with YES confirmation
      --hive-ocm-url string  OCM environment for Hive operations (aliases: production, staging, integration)
`

type replacePullSecretOptions struct {
	clusterID  string
	reason     string
	dryrun     bool
	force      bool
	hiveOcmUrl string
	logger     *logrus.Logger

	genericclioptions.IOStreams
	GlobalOptions *globalflags.GlobalOptions
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

A pre-flight check always runs first. If any checks fail, the command exits
unless --force is specified (requires typing YES to confirm).

See documentation prior to executing:
https://github.com/openshift/ops-sop/blob/master/hypershift/knowledge_base/howto/replace-pull-secret.md
https://github.com/openshift/ops-sop/blob/master/v4/howto/transfer_cluster_ownership.md`,
		Example: `  # Update pull secret on a cluster
  osdctl cluster pull-secret update --cluster-id 1kfmyclusterid --reason "OHSS-1234"

  # Dry-run to preview without making changes
  osdctl cluster pull-secret update --cluster-id 1kfmyclusterid --reason "OHSS-1234" --dry-run

  # Force proceed despite pre-flight failures (e.g. missing pull secret)
  osdctl cluster pull-secret update --cluster-id 1kfmyclusterid --reason "OHSS-1234" --force`,
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
	cmd.Flags().BoolVar(&ops.force, "force", false, "Proceed despite pre-flight failures with YES confirmation")
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
	op := controller.NewPullSecretOp(o.dryrun, logger, out)

	log.SetLogger(zap.New(zap.WriteTo(o.ErrOut), zap.Level(zapcore.WarnLevel)))

	// ================================================================
	// Step 1: OCM data and cluster connectivity
	// ================================================================

	op.Section(1, "OCM data and cluster connectivity",
		"Resolve the cluster, owner account, and OCM access token.",
		"Establish connections to the infrastructure and target clusters.")

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
	o.clusterID = cluster.ID()

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

	fmt.Fprintf(out, "\n  Cluster:  %s (%s)\n", cluster.Name(), o.clusterID)
	fmt.Fprintf(out, "  Type:     %s\n", clusterType)
	fmt.Fprintf(out, "  Owner:    %s (account: %s)\n", ownerUsername, ownerAccountID)
	fmt.Fprintf(out, "  Reason:   %s\n", o.reason)
	if o.dryrun {
		fmt.Fprintf(out, "  Mode:     %s\n", colorDryRun("DRY-RUN (no changes will be made)"))
	}
	if siblingCount > 1 {
		fmt.Fprintf(out, "\n  %s This account owns %d clusters sharing the same access token.\n", colorWarn("[NOTE]"), siblingCount)
		fmt.Fprintf(out, "           This command only updates the pull secret on the cluster above.\n")
		fmt.Fprintf(out, "           Use 'osdctl cluster pull-secret audit -C %s' to review all clusters.\n", o.clusterID)
	}

	fmt.Fprint(out, "\nIs this the correct cluster? ")
	if !utils.ConfirmPrompt() {
		return fmt.Errorf("operation aborted by user")
	}

	// Resolve infrastructure clusters
	var hiveOCM *sdk.Connection
	if o.hiveOcmUrl != "" {
		logger.Infof("Creating separate OCM connection for Hive operations: %s", o.hiveOcmUrl)
		hiveOCM, err = utils.CreateConnectionWithUrl(o.hiveOcmUrl)
		if err != nil {
			op.Fail("could not create hive OCM connection: %v", err)
		} else {
			defer hiveOCM.Close()
			op.OK("hive OCM connection established (%s)", o.hiveOcmUrl)
		}
	}

	var mgmtCluster *cmv1.Cluster
	var masterCluster *cmv1.Cluster

	if isHCP {
		mgmtCluster, err = utils.GetManagementCluster(o.clusterID)
		if err != nil {
			op.Fail("could not resolve management cluster: %v", err)
		} else {
			op.OK("management cluster: %s", mgmtCluster.Name())
		}
		svcCluster, svcErr := utils.GetServiceCluster(o.clusterID)
		if svcErr != nil {
			op.Fail("could not resolve service cluster: %v", svcErr)
		} else {
			masterCluster = svcCluster
			op.OK("service cluster: %s", svcCluster.Name())
		}
	} else {
		var hiveCluster *cmv1.Cluster
		if hiveOCM != nil {
			hiveCluster, err = utils.GetHiveClusterWithConn(o.clusterID, ocm, hiveOCM)
		} else {
			hiveCluster, err = utils.GetHiveCluster(o.clusterID)
		}
		if err != nil {
			op.Fail("could not resolve hive cluster: %v", err)
		} else {
			masterCluster = hiveCluster
			op.OK("hive cluster: %s", hiveCluster.Name())
		}
	}

	// Fetch OCM access token
	var pullSecret []byte
	var auths map[string]*amv1.AccessTokenAuth
	pullSecret, auths, _ = op.FetchAccessTokenOp(ocm, ownerUsername)

	// Connect to clusters
	elevationReasons := []string{
		o.reason,
		"Replacing pull secret using osdctl pull-secret update",
	}

	var masterKubeCli client.Client
	var masterKubeClientSet *kubernetes.Clientset

	if masterCluster != nil {
		logger.Infof("Connecting to infrastructure cluster %s", masterCluster.Name())
		if hiveOCM != nil {
			masterKubeCli, _, masterKubeClientSet, err = common.GetKubeConfigAndClientWithConn(masterCluster.ID(), hiveOCM, elevationReasons...)
		} else {
			masterKubeCli, _, masterKubeClientSet, err = common.GetKubeConfigAndClient(masterCluster.ID(), elevationReasons...)
		}
		if err != nil {
			op.Fail("could not connect to infrastructure cluster %s: %v", masterCluster.Name(), err)
		} else {
			op.OK("connected to infrastructure cluster %s with elevation", masterCluster.Name())
		}
	}

	var targetClientSet *kubernetes.Clientset

	logger.Infof("Connecting to target cluster %s", cluster.Name())
	_, _, targetClientSet, err = common.GetKubeConfigAndClient(o.clusterID, elevationReasons...)
	if err != nil {
		op.Fail("could not connect to target cluster %s: %v", cluster.Name(), err)
	} else {
		op.OK("connected to target cluster %s with elevation", cluster.Name())
	}

	// ================================================================
	// Step 2: Update pull secret on infrastructure cluster
	// ================================================================

	infraName := "(unresolved)"
	if masterCluster != nil {
		infraName = masterCluster.Name()
	}

	if isHCP {
		mgmtName := ""
		if mgmtCluster != nil {
			mgmtName = mgmtCluster.Name()
		}
		op.Section(2, "Update pull secret via ManifestWork (HCP)",
			"HCP clusters store the pull secret inside a ManifestWork on the service cluster.",
			fmt.Sprintf("The ManifestWork %s/%s will be updated with new auth data.", mgmtName, o.clusterID),
			"The work agent then syncs the secret from service cluster → management cluster → hosted cluster.")

		if masterKubeClientSet != nil {
			op.Would("get and update ManifestWork %s/%s on service cluster %s", mgmtName, o.clusterID, infraName)
			op.CheckCanI(ctx, masterKubeClientSet, infraName, "get", "manifestworks", "work.open-cluster-management.io", mgmtName)
			op.CheckCanI(ctx, masterKubeClientSet, infraName, "update", "manifestworks", "work.open-cluster-management.io", mgmtName)
		} else {
			op.Would("get and update ManifestWork on service cluster %s", infraName)
			op.Fail("cannot verify — infrastructure cluster not connected")
		}

		if !o.dryrun && op.AllOK {
			err = updateManifestWork(ocm, masterKubeCli, o.clusterID, mgmtCluster.Name(), pullSecret)
			if err != nil {
				return fmt.Errorf("failed to update pull secret via ManifestWork: %w", err)
			}
			op.OK("ManifestWork updated successfully")
		}
	} else {
		op.Section(2, "Update pull secret via Hive SyncSet (Classic)",
			"Classic clusters store the pull secret in a Hive namespace on the hive cluster.",
			"The old secret is deleted and recreated, then a SyncSet syncs it to the target cluster.",
			"After sync completes, the SyncSet is cleaned up.")

		if masterKubeCli != nil && masterKubeClientSet != nil {
			hiveInfo, found := op.FindHiveNamespaceOp(ctx, masterKubeCli, o.clusterID, infraName)
			if found {
				hiveNS := hiveInfo.Namespace
				op.CheckSecretExists(ctx, masterKubeClientSet, hiveNS, "pull", infraName)

				op.Would("delete and recreate secret %s/pull on %s", hiveNS, infraName)
				op.CheckCanI(ctx, masterKubeClientSet, infraName, "get", "secrets", "", hiveNS)
				op.CheckCanI(ctx, masterKubeClientSet, infraName, "delete", "secrets", "", hiveNS)
				op.CheckCanI(ctx, masterKubeClientSet, infraName, "create", "secrets", "", hiveNS)

				op.Would("create SyncSet %s/pull-secret-replacement to sync to %s", hiveNS, cluster.Name())
				op.CheckCanI(ctx, masterKubeClientSet, infraName, "create", "syncsets", "hive.openshift.io", hiveNS)

				op.Would("poll ClusterSync %s/%s then delete SyncSet", hiveNS, hiveInfo.ClusterDeploymentName)
				op.CheckCanI(ctx, masterKubeClientSet, infraName, "get", "clustersync", "hiveinternal.openshift.io", hiveNS)
				op.CheckCanI(ctx, masterKubeClientSet, infraName, "delete", "syncsets", "hive.openshift.io", hiveNS)
			}
		} else {
			op.Would("resolve Hive namespace, update secret, create SyncSet on %s", infraName)
			op.Fail("cannot verify — infrastructure cluster not connected")
		}

		if !o.dryrun && op.AllOK {
			err = updatePullSecret(ocm, masterKubeCli, masterKubeClientSet, o.clusterID, pullSecret)
			if err != nil {
				return fmt.Errorf("failed to update pull secret via Hive SyncSet: %w", err)
			}
			op.OK("pull secret updated via Hive SyncSet")
		}
	}

	// ================================================================
	// Step 3: Pod rollouts (Classic only)
	// ================================================================

	step := 3
	if !isHCP {
		op.Section(step, "Pod rollouts (Classic only)",
			"After the pull secret is synced, telemeter-client and ocm-agent pods are restarted",
			"so they pick up the new credentials. HCP clusters do not require pod rollouts.")

		if targetClientSet != nil {
			op.Would("roll out pods openshift-monitoring/telemeter-client on %s", cluster.Name())
			op.CheckCanI(ctx, targetClientSet, cluster.Name(), "delete", "pods", "", "openshift-monitoring")

			op.Would("roll out pods openshift-ocm-agent-operator/ocm-agent on %s", cluster.Name())
			op.CheckCanI(ctx, targetClientSet, cluster.Name(), "delete", "pods", "", "openshift-ocm-agent-operator")
		} else {
			op.Fail("cannot verify — target cluster not connected")
		}

		if !o.dryrun && op.AllOK {
			logger.Info("Rolling out pods openshift-monitoring/telemeter-client")
			if err := rolloutPods(targetClientSet, "openshift-monitoring", "app.kubernetes.io/name=telemeter-client"); err != nil {
				op.Warn("failed to roll out telemeter-client pods: %v", err)
			}
		}
		step++
	}

	// ================================================================
	// Step N: Verify pull secret on target cluster
	// ================================================================

	op.Section(step, "Verify pull secret on target cluster",
		"The pull secret on the target cluster is compared against the OCM",
		"access token to verify all auth entries match (token + email per registry).",
		"Required registries are also checked to ensure the cluster can pull images.")

	if targetClientSet != nil {
		op.CheckCanI(ctx, targetClientSet, cluster.Name(), "get", "secrets", "", "openshift-config")

		if auths != nil {
			op.Info("Checking secret openshift-config/pull-secret on %s...", cluster.Name())
			result, verifyErr := controller.VerifyPullSecretAuths(ctx, targetClientSet, auths, out)
			if verifyErr != nil {
				op.Warn("pull secret verification: %v", verifyErr)
			} else if result.Matched == result.Total && o.dryrun {
				op.Info("Pull secret is already up to date — a live run would be a no-op")
			}
		} else {
			op.Fail("cannot compare — OCM access token not available")
		}
	} else {
		op.Fail("cannot verify — target cluster not connected")
	}

	if !isHCP && !o.dryrun && op.AllOK {
		logger.Info("Rolling out pods openshift-ocm-agent-operator/ocm-agent")
		if err := rolloutPods(targetClientSet, "openshift-ocm-agent-operator", "app=ocm-agent"); err != nil {
			op.Warn("failed to roll out ocm-agent pods: %v", err)
		}
	}
	step++

	// ================================================================
	// Step N+1: Service log
	// ================================================================

	op.Section(step, "Send internal service log",
		"An internal (non-customer-visible) service log is sent to record that the",
		"pull secret was updated, including the owner username and reason.")

	op.Would("send internal service log for %s", cluster.Name())

	if !o.dryrun && op.AllOK {
		postCmd := servicelog.PostCmdOptions{
			ClusterId: o.clusterID,
			TemplateParams: []string{
				fmt.Sprintf("MESSAGE=Pull secret replaced for cluster owner '%s'. Reason: %s", ownerUsername, o.reason),
			},
			InternalOnly: true,
		}
		if err := postCmd.Run(); err != nil {
			op.Warn("failed to send internal service log: %v", err)
			fmt.Fprintf(out, "Please manually send: osdctl servicelog post -i %s -p MESSAGE=\"Pull secret replaced for cluster owner '%s'.\"\n", o.clusterID, ownerUsername)
		} else {
			op.OK("internal service log sent")
		}
	}

	// ================================================================
	// Summary
	// ================================================================

	if o.dryrun {
		if op.AllOK {
			fmt.Fprintf(out, "\n%s %s All pre-flight checks passed. No changes were made.\n", colorDryRun("[Dry Run]"), colorOK("OK"))
		} else {
			fmt.Fprintf(out, "\n%s %s Some checks failed. Use --force to proceed despite errors.\n", colorDryRun("[Dry Run]"), colorFail("[FAIL]"))
		}
		return nil
	}

	if !op.AllOK {
		if !o.force {
			fmt.Fprintf(out, "\n%s Pre-flight checks failed. Use --force to proceed despite errors.\n", colorFail("[FAIL]"))
			return fmt.Errorf("pre-flight checks failed — use --force to override")
		}
		fmt.Fprintf(out, "\n%s --force specified. Some checks had failures.\n", colorWarn("[WARN]"))
		fmt.Fprintf(out, "Proceeding may result in errors or incomplete updates.\n")
		fmt.Fprintf(out, "Type YES to confirm you want to proceed: ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil || response != "YES" {
			return fmt.Errorf("operation aborted by user")
		}
	} else {
		fmt.Fprintf(out, "\n%s Pull secret update completed successfully\n", colorOK("[OK]"))
	}

	return nil
}

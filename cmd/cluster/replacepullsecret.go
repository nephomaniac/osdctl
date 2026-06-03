package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"

	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"

	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"

	"github.com/openshift/osdctl/cmd/common"
	"github.com/openshift/osdctl/cmd/servicelog"
	"github.com/openshift/osdctl/internal/utils/globalflags"
	"github.com/openshift/osdctl/pkg/utils"
)

var reasonPattern = regexp.MustCompile(`(?i)(OHSS|PD|SREP|OSD|SDE|ROSAENG)-\d+`)

// requiredPullSecretAuths lists the registry auth entries that must be present
// in a cluster's pull secret for the cluster to function. Missing entries
// indicate an issue with the OCM account or the cluster's pull secret state.
var requiredPullSecretAuths = []string{
	"cloud.openshift.com",
	"quay.io",
	"registry.redhat.io",
	"registry.connect.redhat.com",
}

type replacePullSecretOptions struct {
	clusterID string
	reason    string
	dryrun    bool

	genericclioptions.IOStreams
	GlobalOptions *globalflags.GlobalOptions
}

func newCmdReplacePullSecret(streams genericclioptions.IOStreams, globalOpts *globalflags.GlobalOptions) *cobra.Command {
	ops := &replacePullSecretOptions{
		IOStreams:      streams,
		GlobalOptions: globalOpts,
	}
	cmd := &cobra.Command{
		Use:   "replace-pull-secret",
		Short: "Replace a cluster's pull secret with current OCM access token data",
		Long: `Replace a cluster's pull secret with current OCM access token data.

This updates the pull secret on a ROSA HCP or Classic cluster without performing
an ownership transfer. The pull secret is refreshed using the current cluster
owner's OCM access token.

See documentation prior to executing:
https://github.com/openshift/ops-sop/blob/master/hypershift/knowledge_base/howto/replace-pull-secret.md
https://github.com/openshift/ops-sop/blob/master/v4/howto/transfer_cluster_ownership.md`,
		Example: `  # Replace pull secret on a cluster
  osdctl cluster replace-pull-secret --cluster-id 1kfmyclusterid --reason "OHSS-1234"

  # Dry-run to preview without making changes
  osdctl cluster replace-pull-secret --cluster-id 1kfmyclusterid --reason "OHSS-1234" --dry-run`,
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

	_ = cmd.MarkFlagRequired("cluster-id")
	_ = cmd.MarkFlagRequired("reason")

	return cmd
}

func (o *replacePullSecretOptions) validate() error {
	if !reasonPattern.MatchString(o.reason) {
		fmt.Fprintf(o.ErrOut, "Warning: --reason %q does not appear to contain a ticket ID (e.g. OHSS-1234)\n", o.reason)
		fmt.Fprint(o.Out, "Continue without a valid ticket reference? ")
		if !utils.ConfirmPrompt() {
			return fmt.Errorf("operation aborted — provide a valid --reason")
		}
	}
	return nil
}

func (o *replacePullSecretOptions) run(ctx context.Context) error {
	out := o.Out
	errOut := o.ErrOut

	ocm, err := utils.CreateConnection()
	if err != nil {
		return fmt.Errorf("failed to create OCM client: %w", err)
	}
	defer func() {
		if closeErr := ocm.Close(); closeErr != nil {
			fmt.Fprintf(errOut, "Cannot close the OCM connection: %v\n", closeErr)
		}
	}()

	// --- Phase 1: Identify cluster and confirm with user before expensive operations ---

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

	fmt.Fprintf(out, "\n============================================================\n")
	fmt.Fprintf(out, " Cluster:  %s (%s)\n", cluster.Name(), o.clusterID)
	fmt.Fprintf(out, " Type:     %s\n", clusterType)
	fmt.Fprintf(out, " Owner:    %s\n", ownerUsername)
	fmt.Fprintf(out, " Reason:   %s\n", o.reason)
	if o.dryrun {
		fmt.Fprintf(out, " Mode:     DRY-RUN (no changes will be made)\n")
	}
	fmt.Fprintf(out, "============================================================\n")
	fmt.Fprint(out, "Is this the correct cluster? ")
	if !utils.ConfirmPrompt() {
		return fmt.Errorf("operation aborted by user")
	}

	// --- Phase 2: Resolve infrastructure clusters ---

	var mgmtCluster *cmv1.Cluster
	var masterCluster *cmv1.Cluster

	if isHCP {
		fmt.Fprintln(out, "\nResolving HCP infrastructure clusters...")
		mgmtCluster, err = utils.GetManagementCluster(o.clusterID)
		if err != nil {
			return fmt.Errorf("failed to get management cluster: %w", err)
		}
		svcCluster, err := utils.GetServiceCluster(o.clusterID)
		if err != nil {
			return fmt.Errorf("failed to get service cluster: %w", err)
		}
		masterCluster = svcCluster
		fmt.Fprintf(out, "  Management cluster: %s\n", mgmtCluster.Name())
		fmt.Fprintf(out, "  Service cluster:    %s\n", svcCluster.Name())
	} else {
		fmt.Fprintln(out, "\nResolving Hive cluster...")
		hiveCluster, err := utils.GetHiveCluster(o.clusterID)
		if err != nil {
			return fmt.Errorf("failed to get hive cluster: %w", err)
		}
		masterCluster = hiveCluster
		fmt.Fprintf(out, "  Hive cluster: %s\n", hiveCluster.Name())
	}

	elevationReasons := []string{
		o.reason,
		"Replacing pull secret using osdctl replace-pull-secret",
	}

	// --- Phase 3: Fetch pull secret from OCM ---

	fmt.Fprintln(out, "\nFetching pull secret from OCM...")
	pullSecret, auths, err := fetchOwnerPullSecret(ocm, ownerUsername, out)
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "  Retrieved %d auth entries from OCM access token\n", len(auths))

	// Validate that the OCM access token includes all required registries
	var missingFromOCM []string
	for _, required := range requiredPullSecretAuths {
		if _, ok := auths[required]; !ok {
			missingFromOCM = append(missingFromOCM, required)
		}
	}
	if len(missingFromOCM) > 0 {
		fmt.Fprintf(errOut, "\nWarning: OCM access token is missing required auth entries:\n")
		for _, m := range missingFromOCM {
			fmt.Fprintf(errOut, "  - %s\n", m)
		}
		fmt.Fprintf(errOut, "This may indicate an issue with the cluster owner's OCM account.\n")
		fmt.Fprint(out, "Continue with incomplete pull secret? ")
		if !utils.ConfirmPrompt() {
			return fmt.Errorf("aborted — OCM access token missing required registries: %v", missingFromOCM)
		}
	}

	// --- Phase 4: Pre-flight checks ---

	fmt.Fprintln(out, "\nConnecting to infrastructure cluster...")
	masterKubeCli, _, masterKubeClientSet, err := common.GetKubeConfigAndClient(masterCluster.ID(), elevationReasons...)
	if err != nil {
		return fmt.Errorf("failed to get kube client for infrastructure cluster %s: %w", masterCluster.ID(), err)
	}

	fmt.Fprintln(out, "Connecting to target cluster...")
	_, _, targetClientSet, err := common.GetKubeConfigAndClient(o.clusterID, elevationReasons...)
	if err != nil {
		return fmt.Errorf("failed to get kube client for target cluster %s: %w", o.clusterID, err)
	}

	if err := preflightCheck(ctx, targetClientSet, isHCP, out); err != nil {
		return err
	}

	// --- Phase 5: Apply pull secret update ---

	if !o.dryrun {
		fmt.Fprint(out, "\nProceed with pull secret replacement? ")
		if !utils.ConfirmPrompt() {
			return fmt.Errorf("operation aborted by user")
		}

		fmt.Fprintln(out, "\nApplying pull secret update...")
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
			fmt.Fprintln(out, "Rolling out telemeter-client pods...")
			if err := rolloutPods(targetClientSet, "openshift-monitoring", "app.kubernetes.io/name=telemeter-client"); err != nil {
				fmt.Fprintf(errOut, "Warning: failed to roll out telemeter-client pods: %v\n", err)
			}
		}
	} else {
		fmt.Fprintln(out, "\n[DRY-RUN] Would apply pull secret update — skipping")
		if isHCP {
			fmt.Fprintln(out, "[DRY-RUN] Would update ManifestWork on service cluster")
		} else {
			fmt.Fprintln(out, "[DRY-RUN] Would update pull secret via Hive SyncSet")
			fmt.Fprintln(out, "[DRY-RUN] Would roll out telemeter-client pods")
		}
	}

	// --- Phase 6: Post-operation verification ---

	if !o.dryrun {
		fmt.Fprintln(out, "\nVerifying pull secret on target cluster...")
		if err := verifyPullSecretAuths(ctx, targetClientSet, auths, out, errOut); err != nil {
			return fmt.Errorf("post-operation verification failed: %w", err)
		}

		if !isHCP {
			fmt.Fprintln(out, "Rolling out ocm-agent pods...")
			if err := rolloutPods(targetClientSet, "openshift-ocm-agent-operator", "app=ocm-agent"); err != nil {
				fmt.Fprintf(errOut, "Warning: failed to roll out ocm-agent pods: %v\n", err)
			}
		}
	} else {
		fmt.Fprintln(out, "\n[DRY-RUN] Would verify pull secret on target cluster — skipping")
		if !isHCP {
			fmt.Fprintln(out, "[DRY-RUN] Would roll out ocm-agent pods")
		}
	}

	// --- Phase 7: Service log ---

	postCmd := servicelog.PostCmdOptions{
		ClusterId: o.clusterID,
		TemplateParams: []string{
			fmt.Sprintf("MESSAGE=Pull secret replaced for cluster owner '%s'. Reason: %s", ownerUsername, o.reason),
		},
		InternalOnly: true,
	}
	postCmd.SetDryRun(o.dryrun)
	if err := postCmd.Run(); err != nil {
		fmt.Fprintf(errOut, "Warning: failed to send internal service log: %v\n", err)
		fmt.Fprintf(errOut, "Please manually send: osdctl servicelog post -i %s -p MESSAGE=\"Pull secret replaced for cluster owner '%s'.\"\n", o.clusterID, ownerUsername)
	}

	if o.dryrun {
		fmt.Fprintln(out, "\n[DRY-RUN] Pull secret replacement preview completed (no changes made)")
	} else {
		fmt.Fprintln(out, "\nPull secret replacement completed successfully")
	}
	return nil
}

// fetchOwnerPullSecret retrieves the cluster owner's pull secret from OCM,
// using impersonation if the current OCM user is not the cluster owner.
// Returns the marshaled pull secret bytes and the raw auth map for verification.
func fetchOwnerPullSecret(ocm *sdk.Connection, ownerUsername string, out io.Writer) ([]byte, map[string]*amv1.AccessTokenAuth, error) {
	currentAccountResp, err := ocm.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		fmt.Fprintf(out, "  Warning: could not fetch current account info, will use impersonation: %v\n", err)
	}

	var response *amv1.AccessTokenPostResponse
	if currentAccountResp != nil && currentAccountResp.Body().Username() == ownerUsername {
		fmt.Fprintln(out, "  Current OCM user matches cluster owner, fetching access token directly")
		response, err = ocm.AccountsMgmt().V1().AccessToken().Post().Send()
	} else {
		fmt.Fprintf(out, "  Impersonating cluster owner '%s' to fetch access token\n", ownerUsername)
		response, err = ocm.AccountsMgmt().V1().AccessToken().Post().Impersonate(ownerUsername).Parameter("body", nil).Send()
	}
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch OCM access token: %w", err)
	}

	auths, ok := response.Body().GetAuths()
	if !ok {
		return nil, nil, fmt.Errorf("failed to get auths from access token response — contact SDB if this persists")
	}

	authsMap := map[string]map[string]string{}
	for k, auth := range auths {
		authsMap[k] = map[string]string{
			"auth":  auth.Auth(),
			"email": auth.Email(),
		}
	}

	pullSecret, err := json.Marshal(map[string]map[string]map[string]string{
		"auths": authsMap,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal pull secret: %w", err)
	}

	return pullSecret, auths, nil
}

// preflightCheck validates that the target cluster's pull-secret exists and is readable
// before attempting any mutations.
func preflightCheck(_ context.Context, clientset *kubernetes.Clientset, isHCP bool, out io.Writer) error {
	fmt.Fprintln(out, "\nPre-flight checks...")

	secret, err := clientset.CoreV1().Secrets("openshift-config").Get(context.TODO(), "pull-secret", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("pre-flight: cannot read pull-secret in openshift-config: %w", err)
	}

	if _, ok := secret.Data[".dockerconfigjson"]; !ok {
		return fmt.Errorf("pre-flight: pull-secret exists but is missing .dockerconfigjson key")
	}

	fmt.Fprintln(out, "  [OK] pull-secret exists in openshift-config")
	fmt.Fprintln(out, "  [OK] .dockerconfigjson key present")

	if !isHCP {
		pods, err := clientset.CoreV1().Pods("openshift-monitoring").List(context.TODO(), metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=telemeter-client",
		})
		if err == nil && len(pods.Items) > 0 {
			fmt.Fprintf(out, "  [OK] telemeter-client pods found (%d)\n", len(pods.Items))
		} else if err != nil {
			fmt.Fprintf(out, "  [WARN] could not list telemeter-client pods: %v\n", err)
		}

		pods, err = clientset.CoreV1().Pods("openshift-ocm-agent-operator").List(context.TODO(), metav1.ListOptions{
			LabelSelector: "app=ocm-agent",
		})
		if err == nil && len(pods.Items) > 0 {
			fmt.Fprintf(out, "  [OK] ocm-agent pods found (%d)\n", len(pods.Items))
		} else if err != nil {
			fmt.Fprintf(out, "  [WARN] could not list ocm-agent pods: %v\n", err)
		}
	}

	fmt.Fprintln(out, "  Pre-flight checks passed")
	return nil
}

// verifyPullSecretAuths programmatically compares the OCM access token auths
// against the pull secret deployed on the target cluster. This replaces the
// manual "does this look right?" prompt from verifyClusterPullSecret().
func verifyPullSecretAuths(_ context.Context, clientset *kubernetes.Clientset, expectedAuths map[string]*amv1.AccessTokenAuth, out io.Writer, errOut io.Writer) error {
	pullSecret, err := clientset.CoreV1().Secrets("openshift-config").Get(context.TODO(), "pull-secret", metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pull-secret from target cluster: %w", err)
	}

	if _, ok := pullSecret.Data[".dockerconfigjson"]; !ok {
		return fmt.Errorf("pull-secret is missing .dockerconfigjson key after update")
	}

	// Use the existing getPullSecretTokenAuth from validatepullsecretext.go
	// to extract and compare each auth entry
	var mismatches []string
	matched := 0
	for authKey, expectedAuth := range expectedAuths {
		clusterAuth, err := getPullSecretTokenAuth(authKey, pullSecret)
		if err != nil {
			mismatches = append(mismatches, fmt.Sprintf("  %s: not found in cluster secret (%v)", authKey, err))
			continue
		}

		if clusterAuth.Auth() != expectedAuth.Auth() {
			mismatches = append(mismatches, fmt.Sprintf("  %s: auth token mismatch", authKey))
			continue
		}

		if clusterAuth.Email() != expectedAuth.Email() {
			mismatches = append(mismatches, fmt.Sprintf("  %s: email mismatch (cluster=%q, expected=%q)", authKey, clusterAuth.Email(), expectedAuth.Email()))
			continue
		}

		matched++
	}

	fmt.Fprintf(out, "  Verified %d/%d auth entries match\n", matched, len(expectedAuths))

	if len(mismatches) > 0 {
		fmt.Fprintf(errOut, "\nVerification found %d mismatch(es):\n", len(mismatches))
		for _, m := range mismatches {
			fmt.Fprintln(errOut, m)
		}
		fmt.Fprint(out, "Continue despite mismatches? ")
		if !utils.ConfirmPrompt() {
			return fmt.Errorf("verification failed — %d auth entries did not match", len(mismatches))
		}
	} else {
		fmt.Fprintln(out, "  [OK] All OCM auth entries verified on target cluster")
	}

	// Validate that required registries are present in the cluster's pull secret
	var missingRequired []string
	for _, required := range requiredPullSecretAuths {
		_, err := getPullSecretTokenAuth(required, pullSecret)
		if err != nil {
			missingRequired = append(missingRequired, required)
		}
	}
	if len(missingRequired) > 0 {
		fmt.Fprintf(errOut, "\nWarning: cluster pull secret is missing required registries:\n")
		for _, m := range missingRequired {
			fmt.Fprintf(errOut, "  - %s\n", m)
		}
		fmt.Fprintf(errOut, "The cluster may have issues pulling images or reporting telemetry.\n")
	} else {
		fmt.Fprintln(out, "  [OK] All required registries present in cluster pull secret")
	}

	return nil
}

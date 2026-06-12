package controller

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	sdk "github.com/openshift-online/ocm-sdk-go"
	amv1 "github.com/openshift-online/ocm-sdk-go/accountsmgmt/v1"
	hiveapiv1 "github.com/openshift/hive/apis/hive/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/osdctl/pkg/utils"
)

var (
	psColorOK   = color.New(color.FgGreen).SprintFunc()
	psColorFail = color.New(color.FgRed).SprintFunc()
	psColorWarn = color.New(color.FgYellow).SprintFunc()
)

// RequiredPullSecretAuths lists the registry auth entries that must be present
// in a cluster's pull secret for the cluster to function.
var RequiredPullSecretAuths = []string{
	"cloud.openshift.com",
	"quay.io",
	"registry.redhat.io",
	"registry.connect.redhat.com",
}

// ClusterSummary holds subscription-level data for a cluster owned by an account.
type ClusterSummary struct {
	Name      string
	ID        string
	Status    string
	CreatedAt time.Time
}

// AuthCheckResult holds the outcome of a single registry auth comparison.
type AuthCheckResult struct {
	Registry   string
	Source     string // "access_token" or "registry_credential"
	OK         bool
	TokenMatch bool
	EmailMatch bool
	Email      string
	Detail     string
}

// PullSecretVerifyResult holds the outcome of a per-registry auth comparison.
type PullSecretVerifyResult struct {
	Matched         int
	Total           int
	Mismatches      []string
	AuthResults     []AuthCheckResult
	MissingRequired []string
}

// FetchOwnerPullSecret retrieves the cluster owner's pull secret from OCM,
// using impersonation if the current OCM user is not the cluster owner.
// Returns the marshaled pull secret bytes and the raw auth map for verification.
func FetchOwnerPullSecret(ocm *sdk.Connection, ownerUsername string, logger *logrus.Logger) ([]byte, map[string]*amv1.AccessTokenAuth, error) {
	currentAccountResp, err := ocm.AccountsMgmt().V1().CurrentAccount().Get().Send()
	if err != nil {
		logger.Warnf("Could not fetch current account info, will use impersonation: %v", err)
	}

	var response *amv1.AccessTokenPostResponse
	if currentAccountResp != nil && currentAccountResp.Body().Username() == ownerUsername {
		logger.Info("Current OCM user matches cluster owner, fetching access token directly")
		response, err = ocm.AccountsMgmt().V1().AccessToken().Post().Send()
	} else {
		logger.Infof("Impersonating cluster owner '%s' to fetch access token", ownerUsername)
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

// ValidateRequiredAuths checks that the OCM access token includes all required
// registry auth entries. Returns the list of missing registries.
func ValidateRequiredAuths(auths map[string]*amv1.AccessTokenAuth) []string {
	var missing []string
	for _, required := range RequiredPullSecretAuths {
		if _, ok := auths[required]; !ok {
			missing = append(missing, required)
		}
	}
	return missing
}

// VerifyPullSecretAuths compares OCM access token auths against the pull
// secret on the target cluster. Writes per-registry results to out.
// Returns a PullSecretVerifyResult with match counts and any mismatches.
func VerifyPullSecretAuths(ctx context.Context, clientset *kubernetes.Clientset, expectedAuths map[string]*amv1.AccessTokenAuth, out io.Writer) (*PullSecretVerifyResult, error) {
	pullSecret, err := clientset.CoreV1().Secrets("openshift-config").Get(ctx, "pull-secret", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret openshift-config/pull-secret from target cluster: %w", err)
	}

	if _, ok := pullSecret.Data[".dockerconfigjson"]; !ok {
		return nil, fmt.Errorf("secret openshift-config/pull-secret is missing .dockerconfigjson key")
	}

	result := &PullSecretVerifyResult{Total: len(expectedAuths)}

	for authKey, expectedAuth := range expectedAuths {
		ar := AuthCheckResult{Registry: authKey, Source: "access_token"}

		clusterAuth, err := extractPullSecretAuth(authKey, pullSecret)
		if err != nil {
			result.Mismatches = append(result.Mismatches, authKey)
			ar.Detail = "not found in cluster secret"
			result.AuthResults = append(result.AuthResults, ar)
			continue
		}

		ar.TokenMatch = clusterAuth.auth == expectedAuth.Auth()
		ar.EmailMatch = clusterAuth.email == expectedAuth.Email()
		ar.Email = expectedAuth.Email()

		if !ar.TokenMatch || !ar.EmailMatch {
			result.Mismatches = append(result.Mismatches, authKey)
			details := ""
			if !ar.TokenMatch {
				details += "token mismatch"
			}
			if !ar.EmailMatch {
				if details != "" {
					details += ", "
				}
				details += fmt.Sprintf("email mismatch (cluster=%q, OCM=%q)", clusterAuth.email, expectedAuth.Email())
			}
			ar.Detail = details
			result.AuthResults = append(result.AuthResults, ar)
			continue
		}

		ar.OK = true
		result.Matched++
		result.AuthResults = append(result.AuthResults, ar)
	}

	for _, required := range RequiredPullSecretAuths {
		if _, err := extractPullSecretAuth(required, pullSecret); err != nil {
			result.MissingRequired = append(result.MissingRequired, required)
		}
	}

	// Write human-readable output if a writer is provided
	if out != nil {
		RenderVerifyResult(result, out)
	}

	return result, nil
}

// RenderVerifyResult writes the verification result in human-readable format.
func RenderVerifyResult(result *PullSecretVerifyResult, out io.Writer) {
	mismatchLine := color.New(color.FgYellow, color.Bold).SprintFunc()
	for _, ar := range result.AuthResults {
		if ar.OK {
			fmt.Fprintf(out, "  %s %-40s token=match, email=match (%s)\n", psColorOK("[OK]"), ar.Registry, ar.Email)
		} else {
			fmt.Fprintf(out, "  %s\n", mismatchLine(fmt.Sprintf("[!] %-40s %s", ar.Registry, ar.Detail)))
		}
	}

	fmt.Fprintf(out, "\n  Verified %d/%d auth entries match\n", result.Matched, result.Total)

	if len(result.MissingRequired) > 0 {
		fmt.Fprintf(out, "\n%s cluster pull secret is missing required registries:\n", psColorWarn("[WARN]"))
		for _, m := range result.MissingRequired {
			fmt.Fprintf(out, "  - %s\n", m)
		}
		fmt.Fprintf(out, "The cluster may have issues pulling images or reporting telemetry.\n")
	} else {
		fmt.Fprintf(out, "  %s All required registries present in cluster pull secret\n", psColorOK("[OK]"))
	}
}

// VerifyRegistryCredentials compares OCM registry credentials against the pull
// secret on the target cluster. Registry credentials use a different token
// format (base64-encoded "username:token") than access token auths.
func VerifyRegistryCredentials(ctx context.Context, ocm *sdk.Connection, clientset *kubernetes.Clientset, accountID string, accountEmail string, out io.Writer) (*PullSecretVerifyResult, error) {
	creds, err := utils.GetRegistryCredentials(ocm, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch registry credentials: %w", err)
	}
	if len(creds) == 0 {
		return nil, fmt.Errorf("no registry credentials found for account %s", accountID)
	}

	pullSecret, err := clientset.CoreV1().Secrets("openshift-config").Get(ctx, "pull-secret", metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret openshift-config/pull-secret: %w", err)
	}

	result := &PullSecretVerifyResult{Total: len(creds)}

	for _, cred := range creds {
		registryID := cred.Registry().ID()

		// Resolve registry name from OCM
		regResp, err := ocm.AccountsMgmt().V1().Registries().Registry(registryID).Get().Send()
		if err != nil {
			ar := AuthCheckResult{Registry: registryID, Source: "registry_credential", Detail: fmt.Sprintf("cannot resolve registry: %v", err)}
			result.Mismatches = append(result.Mismatches, registryID)
			result.AuthResults = append(result.AuthResults, ar)
			continue
		}
		regName, _ := regResp.Body().GetName()
		if regName == "" {
			regName = registryID
		}

		ar := AuthCheckResult{Registry: regName, Source: "registry_credential"}

		token, _ := cred.GetToken()
		username, _ := cred.GetUsername()
		if token == "" || username == "" {
			ar.Detail = "missing token or username in OCM registry credential"
			result.Mismatches = append(result.Mismatches, regName)
			result.AuthResults = append(result.AuthResults, ar)
			continue
		}

		clusterAuth, err := extractPullSecretAuth(regName, pullSecret)
		if err != nil {
			ar.Detail = "not found in cluster secret"
			result.Mismatches = append(result.Mismatches, regName)
			result.AuthResults = append(result.AuthResults, ar)
			continue
		}

		// Registry credential tokens are stored as base64("username:token") in the cluster secret
		expectedToken := fmt.Sprintf("%s:%s", username, token)
		clusterTokenDecoded, err := b64.StdEncoding.DecodeString(clusterAuth.auth)
		if err != nil {
			ar.Detail = "failed to decode cluster token"
			result.Mismatches = append(result.Mismatches, regName)
			result.AuthResults = append(result.AuthResults, ar)
			continue
		}

		ar.TokenMatch = expectedToken == string(clusterTokenDecoded)
		ar.EmailMatch = accountEmail == clusterAuth.email
		ar.Email = accountEmail

		if !ar.TokenMatch || !ar.EmailMatch {
			result.Mismatches = append(result.Mismatches, regName)
			details := ""
			if !ar.TokenMatch {
				details += "token mismatch"
			}
			if !ar.EmailMatch {
				if details != "" {
					details += ", "
				}
				details += fmt.Sprintf("email mismatch (cluster=%q, OCM=%q)", clusterAuth.email, accountEmail)
			}
			ar.Detail = details
			result.AuthResults = append(result.AuthResults, ar)
			continue
		}

		ar.OK = true
		result.Matched++
		result.AuthResults = append(result.AuthResults, ar)
	}

	if out != nil {
		RenderVerifyResult(result, out)
	}

	return result, nil
}

// ThreeWayAuthState describes the sync state of a single auth entry across OCM, hive, and target.
type ThreeWayAuthState struct {
	Registry    string
	InOCM       bool
	InHive      bool
	InTarget    bool
	OCMMatchesHive   bool
	OCMMatchesTarget bool
	HiveMatchesTarget bool
}

// ThreeWayComparison holds the full comparison result across OCM, hive, and target.
type ThreeWayComparison struct {
	Auths            []ThreeWayAuthState
	HiveNeedsUpdate  bool
	TargetNeedsSync  bool
	AllInSync        bool
}

// SimpleAuth holds a registry auth's token and email for generic comparison.
type SimpleAuth struct {
	Auth  string
	Email string
}

// AccessTokenToSimple converts access token auths to SimpleAuth map.
func AccessTokenToSimple(auths map[string]*amv1.AccessTokenAuth) map[string]SimpleAuth {
	result := make(map[string]SimpleAuth, len(auths))
	for k, v := range auths {
		result[k] = SimpleAuth{Auth: v.Auth(), Email: v.Email()}
	}
	return result
}

// CompareThreeWay compares pull secret auths across OCM, hive secret, and target cluster secret.
// ocmAuths maps registry name → SimpleAuth with the expected auth/email values.
// hiveData and targetData are the raw .dockerconfigjson bytes from each secret.
func CompareThreeWay(ocmAuths map[string]SimpleAuth, hiveData []byte, targetData []byte) *ThreeWayComparison {
	result := &ThreeWayComparison{AllInSync: true}

	type parsedAuth struct {
		Auth  string `json:"auth"`
		Email string `json:"email"`
	}
	type parsedPS struct {
		Auths map[string]parsedAuth `json:"auths"`
	}

	var hive, target parsedPS
	hiveAuths := make(map[string]parsedAuth)
	targetAuths := make(map[string]parsedAuth)

	if len(hiveData) > 0 {
		if err := json.Unmarshal(hiveData, &hive); err == nil {
			hiveAuths = hive.Auths
		}
	}
	if len(targetData) > 0 {
		if err := json.Unmarshal(targetData, &target); err == nil {
			targetAuths = target.Auths
		}
	}

	// Only compare registries present in the OCM source being checked.
	// Registries in hive/target but not in OCM are outside this source's scope.
	for registry := range ocmAuths {
		state := ThreeWayAuthState{Registry: registry}

		ocmAuth, inOCM := ocmAuths[registry]
		hiveAuth, inHive := hiveAuths[registry]
		targetAuth, inTarget := targetAuths[registry]

		state.InOCM = inOCM
		state.InHive = inHive
		state.InTarget = inTarget

		if inOCM && inHive {
			state.OCMMatchesHive = ocmAuth.Auth == hiveAuth.Auth && ocmAuth.Email == hiveAuth.Email
		}
		if inOCM && inTarget {
			state.OCMMatchesTarget = ocmAuth.Auth == targetAuth.Auth && ocmAuth.Email == targetAuth.Email
		}
		if inHive && inTarget {
			state.HiveMatchesTarget = hiveAuth.Auth == targetAuth.Auth && hiveAuth.Email == targetAuth.Email
		}

		// Determine if sync is needed
		if inOCM && !state.OCMMatchesHive {
			result.HiveNeedsUpdate = true
			result.AllInSync = false
		}
		if inOCM && !state.OCMMatchesTarget {
			result.TargetNeedsSync = true
			result.AllInSync = false
		}
		if inHive && inTarget && !state.HiveMatchesTarget {
			result.AllInSync = false
		}

		result.Auths = append(result.Auths, state)
	}

	return result
}

// RenderThreeWayComparison prints the three-way comparison in a readable format.
// RenderThreeWayComparison prints the three-way comparison in a readable format.
// sourceLabel identifies the OCM source (e.g. "ACCESS TOKEN AUTHS", "REGISTRY CREDENTIAL AUTHS").
func RenderThreeWayComparison(result *ThreeWayComparison, sourceLabel string, out io.Writer) {
	hdr := color.New(color.FgBlue, color.Bold).SprintFunc()

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{sourceLabel, "OCM↔HIVE", "OCM↔TARGET", "HIVE↔TARGET"})
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
	_ = hdr // header color handled by tablewriter

	for _, a := range result.Auths {
		table.Append([]string{
			a.Registry,
			syncStatus(a.InHive, a.OCMMatchesHive),
			syncStatus(a.InTarget, a.OCMMatchesTarget),
			syncStatus(a.InHive && a.InTarget, a.HiveMatchesTarget),
		})
	}
	table.Render()

	fmt.Fprintln(out)
	if result.AllInSync {
		fmt.Fprintf(out, "  %s All sources in sync\n", psColorOK("[OK]"))
	} else {
		if result.HiveNeedsUpdate {
			fmt.Fprintf(out, "  %s Hive secret needs update from OCM\n", psColorWarn("[!]"))
		}
		if result.TargetNeedsSync {
			fmt.Fprintf(out, "  %s Target cluster needs sync from hive\n", psColorWarn("[!]"))
		}
	}
}

func dim(s string) string {
	return color.New(color.FgHiBlack).Sprint(s)
}

func syncStatus(present bool, matches bool) string {
	if !present {
		return color.New(color.FgYellow).Sprint("missing")
	}
	if matches {
		return color.New(color.FgGreen).Sprint("match")
	}
	return color.New(color.FgYellow, color.Bold).Sprint("DIFFERS")
}

// HiveNamespaceInfo holds the resolved Hive namespace and ClusterDeployment name
// for a given cluster.
type HiveNamespaceInfo struct {
	Namespace           string
	ClusterDeploymentName string
}

// FindHiveNamespace discovers the Hive namespace for a cluster by listing
// ClusterDeployments filtered by the api.openshift.com/id label. This avoids
// the fragile uhc-{env}-{clusterID} namespace construction.
func FindHiveNamespace(ctx context.Context, kubeCli client.Client, clusterID string) (*HiveNamespaceInfo, error) {
	if err := hiveapiv1.AddToScheme(kubeCli.Scheme()); err != nil {
		return nil, fmt.Errorf("failed to add hive scheme: %w", err)
	}

	// Try label-based lookup first (fast, targeted)
	cdList := &hiveapiv1.ClusterDeploymentList{}
	labelSelector := client.MatchingLabels{"api.openshift.com/id": clusterID}
	if err := kubeCli.List(ctx, cdList, labelSelector); err == nil && len(cdList.Items) > 0 {
		cd := cdList.Items[0]
		return &HiveNamespaceInfo{
			Namespace:             cd.Namespace,
			ClusterDeploymentName: cd.Name,
		}, nil
	}

	// Fallback: list all ClusterDeployments and match by ClusterMetadata
	allCDs := &hiveapiv1.ClusterDeploymentList{}
	if err := kubeCli.List(ctx, allCDs); err != nil {
		return nil, fmt.Errorf("failed to list ClusterDeployments: %w", err)
	}

	for _, cd := range allCDs.Items {
		if cd.Spec.ClusterMetadata != nil && cd.Spec.ClusterMetadata.ClusterID == clusterID {
			return &HiveNamespaceInfo{
				Namespace:             cd.Namespace,
				ClusterDeploymentName: cd.Name,
			}, nil
		}
	}

	return nil, fmt.Errorf("no ClusterDeployment found for cluster ID %s", clusterID)
}

// PreflightResult holds the outcome of pre-flight checks.
type PreflightResult struct {
	SecretExists bool
	SecretData   []byte // existing .dockerconfigjson content, nil if missing
}

// PreflightCheck validates that the target cluster's pull-secret exists and is
// readable before attempting any mutations. All operations are read-only.
// Returns a PreflightResult so callers can decide whether to create or update.
func PreflightCheck(ctx context.Context, clientset *kubernetes.Clientset, isHCP bool, clusterName string, out io.Writer) (*PreflightResult, error) {
	fmt.Fprintf(out, "\nPre-flight checks on %s...\n", clusterName)
	result := &PreflightResult{}

	secret, err := clientset.CoreV1().Secrets("openshift-config").Get(ctx, "pull-secret", metav1.GetOptions{})
	if err != nil {
		fmt.Fprintf(out, "  %s secret openshift-config/pull-secret not found on %s\n", psColorWarn("[WARN]"), clusterName)
		fmt.Fprintf(out, "  The pull secret can be rebuilt from the owner's OCM access token and registry credentials.\n")
		return result, nil
	}
	result.SecretExists = true

	data, ok := secret.Data[".dockerconfigjson"]
	if !ok {
		fmt.Fprintf(out, "  %s secret openshift-config/pull-secret exists on %s but is missing .dockerconfigjson key\n", psColorWarn("[WARN]"), clusterName)
		return result, nil
	}
	result.SecretData = data

	fmt.Fprintf(out, "  %s secret openshift-config/pull-secret exists on %s\n", psColorOK("[OK]"), clusterName)
	fmt.Fprintf(out, "  %s secret openshift-config/pull-secret has .dockerconfigjson key\n", psColorOK("[OK]"))

	if !isHCP {
		pods, err := clientset.CoreV1().Pods("openshift-monitoring").List(ctx, metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=telemeter-client",
		})
		if err == nil && len(pods.Items) > 0 {
			fmt.Fprintf(out, "  %s pods openshift-monitoring/telemeter-client found on %s (%d)\n", psColorOK("[OK]"), clusterName, len(pods.Items))
		} else if err != nil {
			fmt.Fprintf(out, "  %s could not list pods openshift-monitoring/telemeter-client on %s: %v\n", psColorWarn("[WARN]"), clusterName, err)
		}

		pods, err = clientset.CoreV1().Pods("openshift-ocm-agent-operator").List(ctx, metav1.ListOptions{
			LabelSelector: "app=ocm-agent",
		})
		if err == nil && len(pods.Items) > 0 {
			fmt.Fprintf(out, "  %s pods openshift-ocm-agent-operator/ocm-agent found on %s (%d)\n", psColorOK("[OK]"), clusterName, len(pods.Items))
		} else if err != nil {
			fmt.Fprintf(out, "  %s could not list pods openshift-ocm-agent-operator/ocm-agent on %s: %v\n", psColorWarn("[WARN]"), clusterName, err)
		}
	}

	fmt.Fprintf(out, "  %s Pre-flight checks passed\n", psColorOK("[OK]"))
	return result, nil
}

// CountOwnerClusters returns the number of active clusters owned by the given
// account ID.
func CountOwnerClusters(ocm *sdk.Connection, accountID string, logger *logrus.Logger) int {
	search := fmt.Sprintf("creator.id = '%s' and status != 'Deprovisioned' and status != 'Archived'", accountID)
	resp, err := ocm.AccountsMgmt().V1().Subscriptions().List().
		Search(search).
		Size(1).
		Send()
	if err != nil {
		logger.Debugf("Could not query sibling clusters: %v", err)
		return 0
	}
	return resp.Total()
}

// ListOwnerSubscriptions returns all active subscriptions for the given account ID.
func ListOwnerSubscriptions(ocm *sdk.Connection, accountID string) ([]ClusterSummary, error) {
	search := fmt.Sprintf("creator.id = '%s' and status != 'Deprovisioned' and status != 'Archived'", accountID)
	resp, err := ocm.AccountsMgmt().V1().Subscriptions().List().
		Search(search).
		Size(100).
		Send()
	if err != nil {
		return nil, err
	}

	var clusters []ClusterSummary
	for _, sub := range resp.Items().Slice() {
		name, _ := sub.GetDisplayName()
		clusterID, _ := sub.GetClusterID()
		status, _ := sub.GetStatus()
		createdAt, _ := sub.GetCreatedAt()

		if clusterID == "" {
			continue
		}

		clusters = append(clusters, ClusterSummary{
			Name:      name,
			ID:        clusterID,
			Status:    status,
			CreatedAt: createdAt,
		})
	}

	return clusters, nil
}

// GetLatestCredentialUpdate returns the most recent UpdatedAt time across
// all registry credentials for the given account.
func GetLatestCredentialUpdate(ocm *sdk.Connection, accountID string) (time.Time, error) {
	creds, err := utils.GetRegistryCredentials(ocm, accountID)
	if err != nil {
		return time.Time{}, err
	}

	var latest time.Time
	for _, cred := range creds {
		if updated, ok := cred.GetUpdatedAt(); ok {
			if updated.After(latest) {
				latest = updated
			}
		}
	}
	return latest, nil
}

// pullSecretAuthEntry holds extracted auth data from a cluster pull secret.
type pullSecretAuthEntry struct {
	auth  string
	email string
}

// extractPullSecretAuth extracts an auth entry from a cluster pull secret by registry name.
func extractPullSecretAuth(authID string, secret *corev1.Secret) (*pullSecretAuthEntry, error) {
	dockerConfigJSON, ok := secret.Data[".dockerconfigjson"]
	if !ok {
		return nil, fmt.Errorf("secret is missing .dockerconfigjson key")
	}

	var parsed struct {
		Auths map[string]struct {
			Auth  string `json:"auth"`
			Email string `json:"email"`
		} `json:"auths"`
	}
	if err := json.Unmarshal(dockerConfigJSON, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse pull secret JSON: %w", err)
	}

	entry, found := parsed.Auths[authID]
	if !found {
		return nil, fmt.Errorf("auth '%s' not found in pull secret", authID)
	}

	return &pullSecretAuthEntry{
		auth:  entry.Auth,
		email: entry.Email,
	}, nil
}

// AuthSource indicates where a pull secret auth entry came from.
type AuthSource string

const (
	SourceAccessToken      AuthSource = "access_token"
	SourceRegistryCredential AuthSource = "registry_credential"
	SourceExisting         AuthSource = "existing"
)

// MergedAuth represents a single registry auth entry with its source and merge status.
type MergedAuth struct {
	Registry string
	Auth     string
	Email    string
	Source   AuthSource
}

// AuthConflict represents a registry where access token and registry credential provide different values.
type AuthConflict struct {
	Registry       string
	AccessTokenAuth string
	RegCredAuth    string
}

// MergeResult holds the outcome of merging auth sources.
type MergeResult struct {
	Auths     map[string]MergedAuth
	Conflicts []AuthConflict
	Added     []string // registries added from registry credentials
}

// BuildPullSecretFromSources merges access token auths and registry credential
// auths into a single pull secret, respecting the priority model:
// 1. Access token auths are always applied (primary source)
// 2. Registry credential auths fill in registries not covered by the access token
// 3. Conflicts (same registry, different values) are detected and reported
//
// If existingSecret is non-nil, registries already present in the cluster that
// are not in either OCM source are preserved.
func BuildPullSecretFromSources(
	accessTokenAuths map[string]*amv1.AccessTokenAuth,
	regCreds []*amv1.RegistryCredential,
	ocm *sdk.Connection,
	existingSecret []byte,
	out io.Writer,
) (*MergeResult, []byte, error) {
	result := &MergeResult{
		Auths: make(map[string]MergedAuth),
	}

	// Start with existing auths if we have them (preserves registries not in OCM)
	if len(existingSecret) > 0 {
		var existing struct {
			Auths map[string]struct {
				Auth  string `json:"auth"`
				Email string `json:"email"`
			} `json:"auths"`
		}
		if err := json.Unmarshal(existingSecret, &existing); err == nil {
			for k, v := range existing.Auths {
				result.Auths[k] = MergedAuth{
					Registry: k,
					Auth:     v.Auth,
					Email:    v.Email,
					Source:   SourceExisting,
				}
			}
		}
	}

	// Layer 1: Access token auths (always applied, overwrites existing)
	if accessTokenAuths != nil {
		for k, auth := range accessTokenAuths {
			result.Auths[k] = MergedAuth{
				Registry: k,
				Auth:     auth.Auth(),
				Email:    auth.Email(),
				Source:   SourceAccessToken,
			}
		}
	}

	// Layer 2: Registry credential auths (supplementary)
	for _, cred := range regCreds {
		token, _ := cred.GetToken()
		username, _ := cred.GetUsername()
		if token == "" || username == "" {
			continue
		}

		registryID := cred.Registry().ID()
		regResp, err := ocm.AccountsMgmt().V1().Registries().Registry(registryID).Get().Send()
		if err != nil {
			if out != nil {
				fmt.Fprintf(out, "  %s could not resolve registry %s: %v\n", psColorWarn("[WARN]"), registryID, err)
			}
			continue
		}
		regName, _ := regResp.Body().GetName()
		if regName == "" {
			continue
		}

		regCredAuth := b64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, token)))

		existing, exists := result.Auths[regName]
		if exists && existing.Source == SourceAccessToken {
			// Check for conflict
			if existing.Auth != regCredAuth {
				result.Conflicts = append(result.Conflicts, AuthConflict{
					Registry:       regName,
					AccessTokenAuth: existing.Auth,
					RegCredAuth:    regCredAuth,
				})
			}
			// Access token takes precedence — don't overwrite
			continue
		}

		if !exists || existing.Source == SourceExisting {
			result.Auths[regName] = MergedAuth{
				Registry: regName,
				Auth:     regCredAuth,
				Email:    "", // registry credentials use the account email, set by caller
				Source:   SourceRegistryCredential,
			}
			result.Added = append(result.Added, regName)
		}
	}

	// Marshal the merged pull secret
	authsMap := make(map[string]map[string]string)
	for k, v := range result.Auths {
		authsMap[k] = map[string]string{
			"auth":  v.Auth,
			"email": v.Email,
		}
	}

	pullSecret, err := json.Marshal(map[string]map[string]map[string]string{
		"auths": authsMap,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal merged pull secret: %w", err)
	}

	return result, pullSecret, nil
}

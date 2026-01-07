package hive

import (
	"context"
	"fmt"
	"os"
	"strings"

	ocmConfig "github.com/openshift-online/ocm-common/pkg/ocm/config"
	sdk "github.com/openshift-online/ocm-sdk-go"
	v1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
	configv1 "github.com/openshift/api/config/v1"
	hivev1 "github.com/openshift/hive/apis/hive/v1"
	common "github.com/openshift/osdctl/cmd/common"
	k8s "github.com/openshift/osdctl/pkg/k8s"
	"github.com/openshift/osdctl/pkg/printer"
	"github.com/openshift/osdctl/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// testHiveLoginOptions defines the struct for running health command
// This command requires the ocm API Token https://cloud.redhat.com/openshift/token be available in the OCM_TOKEN env variable.

type testHiveLoginOptions struct {
	clusterID         string
	output            string
	verbose           bool
	awsProfile        string
	hiveOcmConfigPath string
	hiveOcmURL        string
	reason            string
}

// newCmdHealth implements the health command to describe number of running instances in cluster and the expected number of nodes
func newCmdTestHiveLogin() *cobra.Command {
	ops := newtestHiveLoginOptions()
	testHiveLoginCmd := &cobra.Command{
		Use:               "hive-login",
		Short:             "Test Login into both Target Cluster and supporting Hive Cluster.",
		Args:              cobra.NoArgs,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(ops.run())
		},
	}

	testHiveLoginCmd.Flags().BoolVarP(&ops.verbose, "verbose", "", false, "Verbose output")
	testHiveLoginCmd.Flags().StringVarP(&ops.clusterID, "cluster-id", "C", "", "Cluster ID")
	testHiveLoginCmd.Flags().StringVarP(&ops.awsProfile, "profile", "p", "", "AWS Profile")
	testHiveLoginCmd.Flags().StringVar(&ops.hiveOcmConfigPath, "hive-ocm-config", "", "OCM config for hive if different than Cluster")
	testHiveLoginCmd.Flags().StringVar(&ops.hiveOcmURL, "hive-ocm-url", "", "OCM URL for hive if different than Cluster")

	testHiveLoginCmd.MarkFlagRequired("cluster-id")
	return testHiveLoginCmd
}

func newtestHiveLoginOptions() *testHiveLoginOptions {
	return &testHiveLoginOptions{}
}

func (o *testHiveLoginOptions) complete(cmd *cobra.Command, _ []string) error {

	return nil
}

func printDiv() {
	fmt.Printf("\n---------------------------------------------------------------\n\n")
}

func dumpClusterOperators(kubeClient client.Client) error {
	var cos configv1.ClusterOperatorList
	if err := kubeClient.List(context.TODO(), &cos, &client.ListOptions{}); err != nil {
		fmt.Fprintf(os.Stderr, "error fetching cluster operators, err:'%s'\n", err)
		return err
	}
	table := printer.NewTablePrinter(os.Stderr, 20, 1, 1, ' ')
	table.AddRow([]string{"NAME", "AVAILABLE", "PROGRESSING", "DEGRADED"})
	for _, co := range cos.Items {
		var available configv1.ConditionStatus
		var progressing configv1.ConditionStatus
		var degraded configv1.ConditionStatus
		for _, cond := range co.Status.Conditions {
			switch cond.Type {
			case configv1.OperatorAvailable:
				available = cond.Status
			case configv1.OperatorProgressing:
				progressing = cond.Status
			case configv1.OperatorDegraded:
				degraded = cond.Status
			}
		}
		table.AddRow([]string{co.Name, string(available), string(progressing), string(degraded)})
	}
	table.Flush()
	return nil
}

func getClusterDeployment(hiveKubeClient client.Client, clusterID string) (cd hivev1.ClusterDeployment, err error) {
	var cds hivev1.ClusterDeploymentList
	if err := hiveKubeClient.List(context.TODO(), &cds, &client.ListOptions{}); err != nil {
		fmt.Printf("err fetching cluster deployments, err:'%v'", err)
		return cd, err
	}
	var clusterDeployment hivev1.ClusterDeployment
	for _, cd := range cds.Items {
		if strings.Contains(cd.Namespace, clusterID) {
			return cd, nil
		}
	}
	fmt.Printf("Got Hive ClusterDeployment for target cluster:'%s'\n", clusterDeployment.Name)

	return cd, fmt.Errorf("clusterDeployment for cluster:'%s' not found", clusterID)
}

func (o *testHiveLoginOptions) run() error {

	if len(o.hiveOcmURL) > 0 {
		fmt.Printf("Using Hive OCM URL set in args:'%s'\n", o.hiveOcmURL)
	} else {
		o.hiveOcmURL = viper.GetString("hive_ocm_url")
		if len(o.hiveOcmURL) > 0 {
			fmt.Printf("Got Hive OCM URL from viper vars:'%s'\n", o.hiveOcmURL)
		} else {
			fmt.Printf("No 'separate' Hive OCM URL set, using defaults set for target cluster.\n")
		}
	}

	o.reason = "Testing osdctl clients with cluster admin"
	var hiveOCM *sdk.Connection = nil
	var hiveOCMCfg *ocmConfig.Config = nil
	var hiveCluster *v1.Cluster = nil

	fmt.Printf("Building ocm client using legacy functions and env vars...\n")
	ocmClient, err := utils.CreateConnection()
	if err != nil {
		return err
	}
	defer ocmClient.Close()
	cluster, err := utils.GetClusterAnyStatus(ocmClient, o.clusterID)
	if err != nil {
		fmt.Printf("Failed to fetch cluster '%s' from OCM, err:'%v'", o.clusterID, err)
		return err
	}
	clusterID := cluster.ID()
	if o.clusterID != clusterID {
		fmt.Printf("Using internal ID:'%s' for provided cluster:'%s'\n", clusterID, o.clusterID)
		o.clusterID = clusterID
	}

	fmt.Printf("Fetched cluster from OCM:'%s'\n", clusterID)
	printDiv()

	// Test building all the OCM config from a provided file path...
	if len(o.hiveOcmConfigPath) > 0 {
		fmt.Printf("Attempting to build OCM config from provided file path...\n")
		hiveOCMCfg, err = utils.GetOcmConfigFromFilePath(o.hiveOcmConfigPath)
		if err != nil {
			fmt.Printf("Failed to build Hive OCM config from file path:'%s'\n", o.hiveOcmConfigPath)
			return err
		}
	}

	// Test replacing just the OCM URL for an already built config...
	if len(o.hiveOcmURL) > 0 {
		if hiveOCMCfg == nil {
			fmt.Printf("Attempting to build OCM config...\n")
			hiveOCMCfg, err = utils.GetOCMConfigFromEnv()
			if err != nil {
				fmt.Printf("Failed to build OCM config from legacy function\n")
				return err
			}
		}
		hiveOCMCfg.URL = o.hiveOcmURL
	}

	// Test connecting using OCM config...
	if hiveOCMCfg != nil {
		hiveBuilder, err := utils.GetOCMSdkConnBuilderFromConfig(hiveOCMCfg)
		if err != nil {
			fmt.Printf("Failed to create sdk connection builder from hive ocm cfg, err:'%s'\n", err)
			return err
		}
		hiveOCM, err = hiveBuilder.Build()
		//hiveOCM, err = utils.OCMSdkConnFromFilePath(o.hiveOcmConfigPath)
		if err != nil {
			fmt.Printf("Error connecting to OCM env using config at: '%s'\nErr:%v", o.hiveOcmConfigPath, err)
			return err
		}
		fmt.Printf("Built OCM config and connection from provided config inputs\n")
		printDiv()
	}

	// No OCM related config provided, this will test the legacy path(s)...
	if hiveOCM == nil {
		fmt.Println("---- No hive config provided. Using same OCM connections for target cluster and hive ----")
		hiveOCM = ocmClient
		_, err = utils.GetHiveCluster(clusterID)
		if err != nil {
			fmt.Printf("Failed to fetch hive cluster from OCM with legacy function, err:'%v'", err)
			return err
		}
	}

	printDiv()
	hiveCluster, err = utils.GetHiveClusterWithConn(clusterID, ocmClient, hiveOCM)
	if err != nil {
		fmt.Printf("Failed to fetch hive cluster with provided OCM conneciton, err:'%v'", err)
		return err
	}

	fmt.Printf("Got Hive Cluster from OCM:'%s'\n", hiveCluster.ID())
	printDiv()

	fmt.Println("Attempting to create and test Kube Client with k8s.New()...")
	kubeClient, err := k8s.New(clusterID, client.Options{})
	if err != nil {
		return fmt.Errorf("failed to login to cluster:'%s', err: %w", clusterID, err)
	}
	fmt.Printf("Created client connection to target cluster:'%s', '%s'\n", cluster.ID(), cluster.Name())
	// Test an API call to this cluster, dump the cluster operators...
	err = dumpClusterOperators(kubeClient)
	if err != nil {
		return err
	}
	fmt.Println("Create and test Kube Client with k8s.New() - PASS")
	printDiv()

	fmt.Println("Attempting to create and test Kube Client with k8s.NewWithConn()...")
	hiveClient, err := k8s.NewWithConn(hiveCluster.ID(), client.Options{}, hiveOCM)
	if err != nil {
		return fmt.Errorf("failed to login to hive cluster:'%s', err %w", hiveCluster.ID(), err)
	}
	fmt.Printf("Created client connection to HIVE cluster:'%s', '%s'\n", hiveCluster.ID(), hiveCluster.Name())
	// Test an API call to this cluster, dump the cluster operators...
	err = dumpClusterOperators(hiveClient)
	if err != nil {
		return err
	}
	fmt.Println("Create and test Kube Client with k8s.NewWithConn() - PASS")
	printDiv()

	fmt.Println("Attempting to create and test Kube Client with k8s.NewAsBackplaneClusterAdminWithConn()...")
	hiveAdminClient, err := k8s.NewAsBackplaneClusterAdminWithConn(hiveCluster.ID(), client.Options{}, hiveOCM, o.reason)
	if err != nil {
		return fmt.Errorf("failed to login to hive cluster:'%s', err %w", hiveCluster.ID(), err)
	}
	fmt.Printf("Created 'ClusterAdmin' client connection to HIVE cluster:'%s', '%s'\n", hiveCluster.ID(), hiveCluster.Name())
	// Test an elevated API call to this cluster, dump the cluster operators...
	clusterDep, err := getClusterDeployment(hiveAdminClient, clusterID)
	if err != nil {
		return err
	}
	fmt.Printf("Fetched ClusterDeployment:'%s/%s' for cluster:'%s' from HIVE using elevated client\n", clusterDep.Namespace, clusterDep.Name, clusterID)
	fmt.Println("Create and test Kube Client withk8s.NewAsBackplaneClusterAdminWithConn() - PASS")
	printDiv()

	fmt.Printf("Testing non-backplane-admin client, clientSet GetKubeConfigAndClient() for cluster:'%s'\n", clusterID)
	kubeCli, _, kubeClientSet, err := common.GetKubeConfigAndClient(clusterID)
	// Test an API call to this cluster, dump the cluster operators...
	err = dumpClusterOperators(kubeCli)
	if err != nil {
		return err
	}
	nsList, err := kubeClientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("ClientSet list namespaces failed, err:'%v'\n", err)
		return err
	}
	fmt.Printf("Got '%d' namespaces\n", len(nsList.Items))
	fmt.Println("non-bpadmin Create and test Kube Client, Clientset with GetKubeConfigAndClient() - PASS")
	printDiv()

	fmt.Printf("Testing non-backplane-admin client, clientset GetKubeConfigAndClientWithConn for cluster:'%s'\n", clusterID)
	kubeCli, _, kubeClientSet, err = common.GetKubeConfigAndClientWithConn(clusterID, ocmClient)
	// Test an API call to this cluster, dump the cluster operators...
	err = dumpClusterOperators(kubeCli)
	if err != nil {
		return err
	}
	nsList, err = kubeClientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("ClientSet list namespaces failed, err:'%v'\n", err)
		return err
	}
	fmt.Printf("Got '%d' namespaces\n", len(nsList.Items))
	fmt.Println("non-bpadmin Create and test Kube Client, Clientset with GetKubeConfigAndClientWithConn() - PASS")
	printDiv()

	fmt.Printf("Testing backplane-admin client, clientset GetKubeConfigAndClient() for cluster:'%s'\n", clusterID)
	kubeCli, _, kubeClientSet, err = common.GetKubeConfigAndClient(clusterID, o.reason)
	// Test an API call to this cluster, dump the cluster operators...
	err = dumpClusterOperators(kubeCli)
	if err != nil {
		return err
	}
	OpenshiftMonitoringNamespace := "openshift-monitoring"
	podList, err := kubeClientSet.CoreV1().Pods(OpenshiftMonitoringNamespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("ClientSet list 'openshift-monitoring' pods failed, err:'%v'\n", err)
		return err
	}
	fmt.Printf("Got %d pods in namespace:'%s' :\n", len(podList.Items), OpenshiftMonitoringNamespace)
	for i, pod := range podList.Items {
		fmt.Printf("Got pod (%d/%d): '%s/%s' \n", i, len(podList.Items), pod.Namespace, pod.Name)
	}
	fmt.Println("bpadmin Create and test Kube Client, Clientset with GetKubeConfigAndClient() - PASS")
	printDiv()

	fmt.Printf("Testing backplane-admin GetKubeConfigAndClientWithConn() for cluster:'%s'\n", clusterID)
	kubeCli, _, kubeClientSet, err = common.GetKubeConfigAndClientWithConn(clusterID, ocmClient, o.reason)
	// Test an API call to this cluster, dump the cluster operators...
	err = dumpClusterOperators(kubeCli)
	if err != nil {
		return err
	}
	podList, err = kubeClientSet.CoreV1().Pods("openshift-monitoring").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("ClientSet list 'openshift-monitoring' pods failed, err:'%v'\n", err)
		return err
	}
	fmt.Printf("Got %d pods\n", len(podList.Items))
	for i, pod := range podList.Items {
		fmt.Printf("Got pod (%d/%d): '%s/%s' \n", i, len(podList.Items), pod.Namespace, pod.Name)
	}
	fmt.Println("bpadmin Create and test Kube Client, Clientset with GetKubeConfigAndClientWithConn() - PASS")
	printDiv()
	fmt.Printf("Testing GetHiveBPForCluster() hive backplane connection w/o elevation\n")
	hiveBP, err := utils.GetHiveBPForCluster(clusterID, client.Options{}, "", o.hiveOcmURL)
	if err != nil {
		return err
	}
	// Test an API call to this hive cluster, dump the cluster operators...
	err = dumpClusterOperators(hiveBP)
	if err != nil {
		return err
	}
	fmt.Println("Create and test GetHiveBPForCluster() without elevation reason - PASS")
	printDiv()

	fmt.Printf("Testing GetHiveBPForCluster() hive backplane connection w/o elevation\n")
	hiveBP, err = utils.GetHiveBPForCluster(clusterID, client.Options{}, "Testing hive client backplane connections", o.hiveOcmURL)
	if err != nil {
		return err
	}
	// Test an API call to this hive cluster, dump the cluster operators...
	err = dumpClusterOperators(hiveBP)
	if err != nil {
		return err
	}
	// Test an elevated API call to this cluster, dump the cluster operators...
	clusterDep, err = getClusterDeployment(hiveBP, clusterID)
	if err != nil {
		return err
	}
	fmt.Printf("Fetched ClusterDeployment:'%s/%s' for cluster:'%s' from HIVE using elevated client\n", clusterDep.Namespace, clusterDep.Name, clusterID)
	fmt.Println("Create and test GetHiveBPForCluster() with elevation reason - PASS")
	printDiv()
	fmt.Println("All tests Passed")
	return nil
}

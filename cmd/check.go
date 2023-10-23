/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	calicoclient "github.com/projectcalico/api/pkg/client/clientset_generated/clientset"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: checkRunE,
}

func checkRunE(cmd *cobra.Command, args []string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	// create calico client
	calicoClient, err := calicoclient.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error building calico clientset: %s", err.Error())
	}

	globalNetworkPolicies, err := calicoClient.ProjectcalicoV3().GlobalNetworkPolicies().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		klog.Fatalf("Error listing global network policies: %s", err.Error())
		return err
	}
	for _, np := range globalNetworkPolicies.Items {
		klog.Infof("Checking GlobalNetworkPolicy: %s", np.Name)
		if np.Spec.PreDNAT {
			klog.Warningf("GlobalNetworkPolicy %s has PreDNAT enabled, which is not supported by Antrea", np.Name)
			return fmt.Errorf("GlobalNetworkPolicy %s has PreDNAT enabled, which is not supported by Antrea", np.Name)
		}
		if np.Spec.ServiceAccountSelector != "" {
			klog.Warningf("GlobalNetworkPolicy %s has ServiceAccountSelector set, which is not supported by Antrea", np.Name)
			return fmt.Errorf("GlobalNetworkPolicy %s has ServiceAccountSelector set, which is not supported by Antrea", np.Name)
		}
	}

	networkPolicies, err := calicoClient.ProjectcalicoV3().NetworkPolicies("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		klog.Fatalf("Error listing network policies: %s", err.Error())
		return err
	}
	for _, np := range networkPolicies.Items {
		klog.Infof("Checking NetworkPolicy: %s", np.Name)
		if np.Spec.ServiceAccountSelector != "" {
			klog.Warningf("NetworkPolicy %s has ServiceAccountSelector set, which is not supported by Antrea", np.Name)
			return fmt.Errorf("NetworkPolicy %s has ServiceAccountSelector set, which is not supported by Antrea", np.Name)
		}
	}

	klog.Infof("Calico NetworkPolicy check passed, all Network Polices and Global Network Policies are supported by Antrea")
	return nil
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

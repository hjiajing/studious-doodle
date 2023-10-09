/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	calicoclient "github.com/projectcalico/api/pkg/client/clientset_generated/clientset"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// migrateIPPoolCmd represents the migrateIppool command
var migrateIPPoolCmd = &cobra.Command{
	Use:   "migrate-ippool",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: migrateIPPoolRunE,
}

func migrateIPPoolRunE(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return cmd.Help()
	}
	clusterCIDR := args[2]
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Errorf("Failed to build config, error:", err)
		return err
	}
	calicoClient, err := calicoclient.NewForConfig(config)
	if err != nil {
		klog.Errorf("Failed to create Calico client, error: ", err)
		return err
	}
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("Failed to create Kubernetes client, error:", err)
		return err
	}
	currentIPPoolName := args[0]
	currentIPPool, err := calicoClient.ProjectcalicoV3().IPPools().Get(context.TODO(), currentIPPoolName, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get current IPPool, error:", err)
		return err
	}
	newIPPool := &v3.IPPool{
		ObjectMeta: metav1.ObjectMeta{
			Name: "new-ippool",
		},
		Spec: v3.IPPoolSpec{
			AllowedUses:  currentIPPool.Spec.AllowedUses,
			BlockSize:    currentIPPool.Spec.BlockSize,
			CIDR:         currentIPPool.Spec.CIDR,
			IPIPMode:     currentIPPool.Spec.IPIPMode,
			NATOutgoing:  currentIPPool.Spec.NATOutgoing,
			NodeSelector: currentIPPool.Spec.NodeSelector,
			VXLANMode:    currentIPPool.Spec.VXLANMode,
		},
	}
	fmt.Println(newIPPool.Name)
	return nil
}

func init() {
	rootCmd.AddCommand(migrateIPPoolCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateIppoolCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrateIppoolCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

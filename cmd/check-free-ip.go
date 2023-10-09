/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"net"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// checkFreeIpCmd represents the checkFreeIp command
// antrea-migrator check-free-ip ${CLUSTER_CIDR}
var checkFreeIpCmd = &cobra.Command{
	Use:   "check-free-ip",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: checkFreeIPRunE,
}

// Check if there're enough free IPs for calico IPPool migration.
func checkFreeIPRunE(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Help()
	}
	clusterCIDR := args[0]
	ipNumbers, err := calculateIPs(clusterCIDR)
	if err != nil {
		return err
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Errorf("Failed to build config, error:", err)
		return err
	}

	// create the clientset
	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf("Failed to create Kubernetes client, error:", err)
		return err
	}
	nodes, err := k8sClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list Nodes, error:", err)
		return err
	}

	for _, node := range nodes.Items {
		podCIDR := node.Spec.PodCIDR
		nodeIPNumber, _ := calculateIPs(podCIDR)
		ipNumbers -= nodeIPNumber
	}
	// Get all Pods with Cluster Network
	allPods, err := k8sClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list Pods, error:", err)
		return err
	}
	if ipNumbers > len(allPods.Items) {
		fmt.Printf("There are enough free IPs for calico IPPool migration, the number of free IPs is %d.\n", ipNumbers)
	} else {
		fmt.Printf("There are not enough free IPs for calico IPPool migration, the number of free IPs is %d.\n", ipNumbers)
	}
	return nil
}

// calculateIPs calculates the number of IPs in the given CIDR.
func calculateIPs(cidr string) (int, error) {
	_, podCIDR, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, err
	}
	// Get the mask of the CIDR
	ones, bits := podCIDR.Mask.Size()
	// Calculate the number of IPs in the CIDR
	ipNumbers := 1 << uint(bits-ones)
	return ipNumbers, nil
}

func init() {
	rootCmd.AddCommand(checkFreeIpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkFreeIpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkFreeIpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

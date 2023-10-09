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

// checkNodesCmd represents the checkNodes command
var checkNodesCmd = &cobra.Command{
	Use:   "check-nodes",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: checkNodesRunE,
}

func init() {
	rootCmd.AddCommand(checkNodesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkNodesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkNodesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func checkNodesRunE(cmd *cobra.Command, args []string) error {
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

	noCoexistCNINodes := []string{}
	pods, err := k8sClient.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorln("Failed to create Kubernetes client, error:", err)
		return err
	}
	// ipamblock is a private resource in Calico. So we have to check Pods to get the potential conflict
	// static route.
	// For example:
	// If one Pod's IP address is 10.10.4.8 on Node test-node5, and PodCIDR of test-node2 is 10.10.4.0/24.
	// Then the route to 10.10.4.0 may be conflict. So we should make Node test-node5 offline.
	for _, node := range nodes.Items {
		var nodePodCIDR *net.IPNet
		if node.Spec.PodCIDR != "" {
			_, nodePodCIDR, _ = net.ParseCIDR(node.Spec.PodCIDR)
			klog.Infof("PodCIDR on Node %s is %s\n", node.Name, nodePodCIDR.String())
		} else {
			klog.Infof("Node %s has not been allocated Pod CIDR\n", node.Name)
			continue
		}
		klog.Info("Checking Pod CIDR on Node ", node.Name)
		for _, pod := range pods.Items {
			if nodePodCIDR.Contains(net.ParseIP(pod.Status.PodIP)) {
				noCoexistCNINodes = append(noCoexistCNINodes, pod.Spec.NodeName)
				klog.Infof("PodCIDR intersects on Node %s", pod.Spec.NodeName)
				break
			}
		}
	}

	fmt.Println(noCoexistCNINodes)
	return nil
}

/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"strings"

	"antrea.io/antrea/pkg/apis/controlplane"
	"antrea.io/antrea/pkg/apis/crd/v1alpha1"
	crdclientset "antrea.io/antrea/pkg/client/clientset/versioned"
	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	calicoclient "github.com/projectcalico/api/pkg/client/clientset_generated/clientset"
	"github.com/projectcalico/api/pkg/lib/numorstring"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	defaultPriority = 10
)

// convertNetworkpolicyCmd represents the convertNetworkpolicy command
var convertNetworkpolicyCmd = &cobra.Command{
	Use:   "convert-networkpolicy",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: convertNetworkPolicyRunE,
}

type networkPolicyConverter struct {
	antreaClient *crdclientset.Clientset
	calicoClient *calicoclient.Clientset
}

func newNetworkPolicyConverter(config *restclient.Config) (*networkPolicyConverter, error) {
	calicoClient, err := calicoclient.NewForConfig(config)
	if err != nil {
		klog.Errorf("Failed to create Calico client, error: ", err)
		return nil, err
	}
	antreaClient, err := crdclientset.NewForConfig(config)
	if err != nil {
		klog.Errorf("Failed to create Antrea client, error: ", err)
		return nil, err
	}

	return &networkPolicyConverter{
		antreaClient: antreaClient,
		calicoClient: calicoClient,
	}, nil
}

func convertNetworkPolicyRunE(cmd *cobra.Command, _ []string) error {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Errorf("Failed to build config, error:", err)
		return err
	}
	klog.Info("Converting Namespaced NetworkPolicy")
	npConverter, err := newNetworkPolicyConverter(config)
	if err != nil {
		return err
	}
	klog.Info("Converting Global NetworkPolicy")
	if err := npConverter.convertGlobalNetworkPolicy(); err != nil {
		return err
	}

	return nil
}

func (npConvert *networkPolicyConverter) convertGlobalNetworkPolicy() error {
	globalNetworkPolicies, err := npConvert.calicoClient.ProjectcalicoV3().GlobalNetworkPolicies().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list GlobalNetworkPolicies, error: ", err)
		return err
	}

	for _, np := range globalNetworkPolicies.Items {
		antreaClusterNP, err := convertToAntreaClusterNP(&np)
		if err != nil {
			return err
		}
		klog.InfoS("Creating Antrea Cluster NetworkPolicy", "ClusterNetworkPolicy", antreaClusterNP.Name)
		if _, err := npConvert.antreaClient.CrdV1alpha1().ClusterNetworkPolicies().Create(context.TODO(), antreaClusterNP, metav1.CreateOptions{}); err != nil {
			klog.ErrorS(err, "Failed to create Antrea Cluster NetworkPolicy", "ClusterNetworkPolicy", antreaClusterNP.Name)
			return err
		}
	}
	return nil
}

func (npConvert *networkPolicyConverter) convertNamespacedNetworkPolicy() error {
	networkPolicies, err := npConvert.calicoClient.ProjectcalicoV3().NetworkPolicies("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		klog.Errorf("Failed to list NetworkPolicies, error: ", err)
		return err
	}
	for _, np := range networkPolicies.Items {
		antreaNP := convertToAntreaNP(&np)
		klog.InfoS("Creating Antrea NetworkPolicy", "NetworkPolicy", antreaNP.Name)
		if _, err := npConvert.antreaClient.CrdV1alpha1().NetworkPolicies(antreaNP.Namespace).Create(context.TODO(), antreaNP, metav1.CreateOptions{}); err != nil {
			klog.ErrorS(err, "Failed to create Antrea NetworkPolicy", "NetworkPolicy", antreaNP.Name)
			return err
		}
	}
	return nil
}

func convertToAntreaNP(calicoNP *v3.NetworkPolicy) *v1alpha1.NetworkPolicy {
	antreaNP := &v1alpha1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      calicoNP.Name,
			Namespace: calicoNP.Namespace,
		},
	}
	return antreaNP
}

func convertToAntreaClusterNP(calicoGlobalNP *v3.GlobalNetworkPolicy) (*v1alpha1.ClusterNetworkPolicy, error) {
	clusterNP := &v1alpha1.ClusterNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: calicoGlobalNP.Name,
		},
		Spec: v1alpha1.ClusterNetworkPolicySpec{
			Priority: 10,
		},
	}
	if calicoGlobalNP.Spec.Selector != "" {
		clusterNP.Spec.AppliedTo = []v1alpha1.AppliedTo{
			{
				PodSelector:       selectorToLabelSelector(calicoGlobalNP.Spec.Selector),
				NamespaceSelector: selectorToLabelSelector(calicoGlobalNP.Spec.NamespaceSelector),
			},
		}
	}
	if len(calicoGlobalNP.Spec.Ingress) != 0 {
		for _, ingress := range calicoGlobalNP.Spec.Ingress {
			rule, err := convertToAntreaRule(ingress)
			if err != nil {
				return nil, err
			}
			clusterNP.Spec.Ingress = append(clusterNP.Spec.Ingress, rule)
		}
	}
	if len(calicoGlobalNP.Spec.Egress) != 0 {
		for _, egress := range calicoGlobalNP.Spec.Egress {
			rule, err := convertToAntreaRule(egress)
			if err != nil {
				return nil, err
			}
			clusterNP.Spec.Egress = append(clusterNP.Spec.Egress, rule)
		}
	}
	return clusterNP, nil
}

func convertToAntreaRule(calicoRule v3.Rule) (v1alpha1.Rule, error) {
	rule := v1alpha1.Rule{
		Protocols: make([]v1alpha1.NetworkPolicyProtocol, 0),
	}
	rule.Action = convertAction(calicoRule.Action)
	//rule.Ports = convertProtocol(calicoRule.Protocol)
	rule.From = convertEntityRule(calicoRule.Source)
	rule.To = convertEntityRule(calicoRule.Destination)
	var err error
	rule.Ports, err = convertPort(calicoRule.Source.Ports, calicoRule.Protocol)
	if err != nil {
		return rule, err
	}
	return rule, nil
}

// convertPort converts Calico port to Antrea port
func convertPort(ports []numorstring.Port, protocol *numorstring.Protocol) ([]v1alpha1.NetworkPolicyPort, error) {
	networkPolicyPort := []v1alpha1.NetworkPolicyPort{}
	if len(ports) == 0 {
		if protocol == nil {
			return nil, nil
		}
		// TODO: Numbers are not supported yet
		if protocol.Type == numorstring.NumOrStringString {
			switch protocol.StrVal {
			case numorstring.ProtocolTCP:
				networkPolicyPort = append(networkPolicyPort,
					v1alpha1.NetworkPolicyPort{
						Protocol: toPtr(v1.Protocol(controlplane.ProtocolTCP)),
					})
			case numorstring.ProtocolSCTP:
				networkPolicyPort = append(networkPolicyPort,
					v1alpha1.NetworkPolicyPort{
						Protocol: toPtr(v1.Protocol(controlplane.ProtocolSCTP)),
					})
			case numorstring.ProtocolUDP:
				networkPolicyPort = append(networkPolicyPort,
					v1alpha1.NetworkPolicyPort{
						Protocol: toPtr(v1.Protocol(controlplane.ProtocolUDP)),
					})
			default:
				klog.Warning("Unsupported protocol type: ", protocol.StrVal)
				return nil, fmt.Errorf("unsupported protocol type: %s", protocol.StrVal)
			}
		}
		return networkPolicyPort, nil
	}
	for _, port := range ports {
		antreaPort := v1alpha1.NetworkPolicyPort{
			Protocol: toPtr(v1.Protocol(protocol.StrVal)),
			Port: &intstr.IntOrString{
				Type:   0,
				IntVal: int32(port.MinPort),
			},
			EndPort: toPtr(int32(port.MaxPort)),
		}
		networkPolicyPort = append(networkPolicyPort, antreaPort)
	}
	return networkPolicyPort, nil
}

// convertSource converts Calico source to Antrea source, Calico destination to Antrea destination
func convertEntityRule(source v3.EntityRule) []v1alpha1.NetworkPolicyPeer {
	networkPolicyPeer := make([]v1alpha1.NetworkPolicyPeer, 0)
	for _, net := range source.Nets {
		networkPolicyPeer = append(networkPolicyPeer,
			v1alpha1.NetworkPolicyPeer{
				IPBlock: &v1alpha1.IPBlock{
					CIDR: net,
				},
			})
	}
	networkPolicyPeer = append(networkPolicyPeer,
		v1alpha1.NetworkPolicyPeer{
			PodSelector:       selectorToLabelSelector(source.Selector),
			NamespaceSelector: selectorToLabelSelector(source.NamespaceSelector),
		})
	return networkPolicyPeer
}

// convertAction converts Calico action to Antrea action
func convertAction(action v3.Action) *v1alpha1.RuleAction {
	switch action {
	case v3.Allow:
		return toPtr(v1alpha1.RuleActionAllow)
	case v3.Deny:
		return toPtr(v1alpha1.RuleActionDrop)
	case v3.Pass:
		return toPtr(v1alpha1.RuleActionPass)
	case v3.Log:
		return toPtr(v1alpha1.RuleActionPass)
	}
	return nil
}

// TODO: More complex selector
func selectorToLabelSelector(selector string) *metav1.LabelSelector {
	kv := strings.Split(selector, " == ")
	if len(kv) != 2 {
		return nil
	}
	return &metav1.LabelSelector{
		MatchLabels: map[string]string{
			remoteQuotes(kv[0]): remoteQuotes(kv[1]),
		},
	}
}

// remoteQuotes removes single quotes from the beginning and end of a string.
func remoteQuotes(s string) string {
	if len(s) > 1 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1]
	}
	return s
}

// a genericity function return a pointer to the right value
func toPtr[T any](p T) *T {
	return &p
}

func init() {
	rootCmd.AddCommand(convertNetworkpolicyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// convertNetworkpolicyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// convertNetworkpolicyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

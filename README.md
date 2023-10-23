# Migrate Calico to Antrea

This document is a guide to migrate a Calico Cluster to Antrea Cluster. During the migration, the Service may not be available.

## Terms

+ Calico Cluster: A Kubernetes Cluster with Calico CNI.
+ Antrea Cluster: A Kubernetes Cluster with Antrea CNI.
+ Calico Pod: A Pod with Calico CNI.
+ Antrea Pod: A Pod with Antrea CNI.

## Steps

1. Build antrea-migrator, which is a tool for Network Policy converting. It can also check if the requirement is satisfied.
2. Run antrea-migrator to check if the requirement is satisfied.
3. Deploy Antrea to the cluster, now that there are two CNI in the cluster. In this step, all legacy Pods are still using Calico CNI, but the new Pods will use Antrea 
   CNI. **NOTE: The Calico Pods cannot connect to the Antrea Pods. The Service in the Cluster might be unavailable in this step**
4. Run antrea-migrator to convert Calico Networkpolicy to Antrea Networkpolicy. 
   The Calico GlobalNetworkPolicy is converted to Antrea ClusterNetworkPolicy and Calico NetworkPolicy is converted to Antrea NetworkPolicy.
5. Run a sandbox killer job on every Node, this job will kill all `/pause` process on the Node, so that all Pods on the Node
   will restart in place. The restart event will trigger a CNI event to use Antrea CNI. After this step, all Pods in the
   Cluster will be Antrea Pods.
6. Remove Calico CNI, all calico network interfaces and corresponding static IP routes will be removed.
7. Remove legacy iptables rules of Calico.

## Example

### Build antrea-migrator

```Bash
$ GOOS=linux GOARCH=amd64 go build -o antrea-migrator ./main.go
```

### Check the requirement

NOT all Calico NetworkPolicy is supported by Antrea. The antrea-migrator can check if the NetworkPolicy is supported by Antrea.
If the requirement is not satisfied, the antrea-migrator will print the unsupported NetworkPolicy, and stop the migration.

```Bash
$ ./antrea-migrator print-requirements
========== WARNING ==========
THIS IS AN EXPERIMENTAL FEATURE.
YOUR SERVICE MAY NOT WORK AS EXPECTED DURING THE MIGRATION.
IF YOU WANT TO MIGRATE YOUR SERVICE, PLEASE MAKE SURE THE CALICO APISERVER IS INSTALLED.
$ ./antrea-migrator check
I1023 05:35:50.581556 1291848 check.go:47] Checking GlobalNetworkPolicy: deny-blue
I1023 05:35:50.581634 1291848 check.go:47] Checking GlobalNetworkPolicy: deny-green
I1023 05:35:50.581650 1291848 check.go:47] Checking GlobalNetworkPolicy: deny-nginx-ds
I1023 05:35:50.592310 1291848 check.go:71] Calico NetworkPolicy check passed, all Network Polices and Global Network Policies are supported by Antrea
```

### Deploy Antrea

Deploy Antrea to the cluster, now that there are two CNI in the cluster. In this step, all legacy Pods are still using Calico CNI, but the new Pods will use Antrea. The Calico Pods could still communicate with each other, but the Calico Pods cannot connect to the Antrea Pods. The Service in the Cluster might be unavailable in this step.

But sometimes the Pod IP might be conflict with `antrea-gw0`'s IP. In this case, the Pod will be unaccessible. You can check the Pod's IP and `antrea-gw0`'s IP to see if there is a conflict.

### Convert NetworkPolicy

This step will convert Calico NetworkPolicy to Antrea NetworkPolicy. The Calico GlobalNetworkPolicy is converted to Antrea ClusterNetworkPolicy and Calico NetworkPolicy is converted to Antrea NetworkPolicy.

`antrea-migrator` will list and traverse all Calico NetworkPolicy and GlobalNetworkPolicy, and convert them to Antrea NetworkPolicy and ClusterNetworkPolicy. The result may contain several warnnings.

```Bash
./antrea-migrator convert-networkpolicy
I1023 06:13:18.813386 1721989 convert-networkpolicy.go:72] Converting Namespaced NetworkPolicy
I1023 06:13:18.814819 1721989 convert-networkpolicy.go:77] Converting Global NetworkPolicy
I1023 06:13:18.920224 1721989 convert-networkpolicy.go:97] "Creating Antrea Cluster NetworkPolicy" ClusterNetworkPolicy="deny-blue"
W1023 06:13:18.962208 1721989 warnings.go:70] unknown field "spec.ingress[0].to"
I1023 06:13:18.962989 1721989 convert-networkpolicy.go:97] "Creating Antrea Cluster NetworkPolicy" ClusterNetworkPolicy="deny-green"
W1023 06:13:18.977663 1721989 warnings.go:70] unknown field "spec.ingress[0].to"
I1023 06:13:18.977908 1721989 convert-networkpolicy.go:97] "Creating Antrea Cluster NetworkPolicy" ClusterNetworkPolicy="deny-nginx-ds"
W1023 06:13:18.992997 1721989 warnings.go:70] unknown field "spec.ingress[0].to"

# Check the Network Policy
$ kubectl get globalnetworkpolicies.crd.projectcalico.org
NAME                    AGE
default.deny-blue       3m29s
default.deny-green      3m26s
default.deny-nginx-ds   3m22s
$ kubectl get clusternetworkpolicies.crd.antrea.io
NAME            TIER          PRIORITY   DESIRED NODES   CURRENT NODES   AGE
deny-blue       application   10         1               1               41s
deny-green      application   10         0               0               41s
deny-nginx-ds   application   10         1               1               41s
```

### Kill sandbox

This step will kill all `/pause` process on the Node, so that all Pods on the Node will restart in place. The restart event will trigger a CNI event to use Antrea CNI. After this step, all Pods in the Cluster will be Antrea Pods.

During this step, several Jobs will be created, whose name is `antrea-migrator-sandbox-killer-<NodeName>`. The Job will be deleted after the Pod is killed. The Job use `spec.nodeName` to select the Node, so that the Job will be scheduled to the Node.

While this, the workload Pods' statues could be `comepeled` or `CrashLoopBackOff`, and the Service related to the workload Pods could be unavailable. But the Pods will restart soon because they do not need to be rescheduled and pull the image again.

```Bash
    cat <<EOF | kubectl apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: antrea-migrator-kill-sandbox-$node
  namespace: kube-system
spec:
    template:
      spec:
        nodeName: $node
        hostPID: true
        containers:
        - name: antrea-migrator-kill-sandbox
          image: busybox:1.28
          command: 
          - "/bin/sh"
          - "-c"
          - "pkill -9 /pause"
          securityContext:
            privileged: true
        restartPolicy: Never
EOF
========== Creating Sandbox killer job on Node: kind-control-plane ==========
job.batch/antrea-migrator-kill-sandbox-kind-control-plane unchanged
========== Creating Sandbox killer job on Node: kind-worker ==========
job.batch/antrea-migrator-kill-sandbox-kind-worker unchanged
========== Creating Sandbox killer job on Node: kind-worker10 ==========
job.batch/antrea-migrator-kill-sandbox-kind-worker10 unchanged
job.batch/antrea-migrator-kill-sandbox-kind-worker6 created
========== Creating Sandbox killer job on Node: kind-worker8 ==========
job.batch/antrea-migrator-kill-sandbox-kind-worker8 created
========== Creating Sandbox killer job on Node: kind-worker9 ==========
job.batch/antrea-migrator-kill-sandbox-kind-worker9 created
========== Waiting for Sandbox killer job on Node: kind-control-plane ==========
job.batch/antrea-migrator-kill-sandbox-kind-control-plane condition met
========== Waiting for Sandbox killer job on Node: kind-worker ==========
job.batch/antrea-migrator-kill-sandbox-kind-worker condition met
========== Waiting for Sandbox killer job on Node: kind-worker10 ==========
job.batch/antrea-migrator-kill-sandbox-kind-worker10 condition met
========== Waiting for Sandbox killer job on Node: kind-worker8 ==========
job.batch/antrea-migrator-kill-sandbox-kind-worker8 condition met
========== Waiting for Sandbox killer job on Node: kind-worker9 ==========
job.batch/antrea-migrator-kill-sandbox-kind-worker9 condition met
```

### Remove Calico CNI

This step will remove Calico CNI by manifest, all calico network interfaces and corresponding static IP routes will be removed.

```Bash
kubectl delete -f https://docs.projectcalico.org/manifests/calico.yaml
```

### Remove legacy iptables rules of Calico

Although the Calico CNI is removed, the legacy iptables rules of Calico still exist. This step will remove the legacy iptables rules of Calico.

```Bash
antrea_agents=$(kubectl get pods -n kube-system -l app=antrea -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')
for antrea_agent in $antrea_agents; do
    echo "========== Removing Calico iptables rules on Node: $antrea_agent =========="
    kubectl exec  $antrea_agent -n kube-system -- /bin/sh -c "iptables-save | grep -v cali | iptables-restore"
done
```
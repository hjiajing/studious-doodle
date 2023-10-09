# Migrate CNI from Calico to Antrea

This document describes how to migrate CNI from Calico to Antrea.

## Terms

+ Calico Pod: A Pod managed by Calico CNI.
+ Antrea Pod: A Pod managed by Antrea CNI.
+ Multus Pod: A Pod managed by Multus CNI.

## Pre-test

### Deploy Antrea directly

The CNI configuration files are located at `/etc/cni/net.d/`. If you have installed Calico, you can find the
configuration files of Calico CNI. If you have installed Antrea, you can find the configuration files of Antrea CNI.

In Kubernetes, the Kubelet use CNI configuration file by alphabetical order. So if we deploy both Calico and Antrea, the
Kubelet will use Antrea CNI configuration file (`/etc/cni/net.d/10-antrea.conflist`). Any Pod created after
Antrea installation will use Antrea CNI. Any Pod created before Antrea installation will use Calico CNI. In this case,
the Antrea Pod could communicate with Calico Pod, but the Calico Pod could not communicate with Antrea Pod (A few Calico
Pods could communicate with Antrea Pod, but most of them could not).

For example:

```bash
# These Pods are Calico Pods
❯ kubectl get pod -o wide
NAME             READY   STATUS    RESTARTS   AGE   IP              NODE           NOMINATED NODE   READINESS GATES
nginx-ds-5266z   1/1     Running   0          82s   10.10.39.193    test-worker2   <none>           <none>
nginx-ds-97nlh   1/1     Running   0          81s   10.10.72.1      test-worker4   <none>           <none>
nginx-ds-bmv2l   1/1     Running   0          82s   10.10.244.69    test-worker8   <none>           <none>
nginx-ds-j47sp   1/1     Running   0          80s   10.10.8.129     test-worker5   <none>           <none>
nginx-ds-j7mlw   1/1     Running   0          82s   10.10.129.193   test-worker7   <none>           <none>
nginx-ds-pqd27   1/1     Running   0          81s   10.10.222.65    test-worker    <none>           <none>
nginx-ds-rj24g   1/1     Running   0          80s   10.10.4.129     test-worker3   <none>           <none>
nginx-ds-zstmr   1/1     Running   0          82s   10.10.247.65    test-worker6   <none>           <none>
# Deploy Antrea and restart one nginx Pod
❯ kubectl delete pod nginx-ds-5266z
pod "nginx-ds-5266z" deleted
❯ kubectl get pod -o wide
NAME             READY   STATUS    RESTARTS   AGE     IP              NODE           NOMINATED NODE   READINESS GATES
nginx-ds-97nlh   1/1     Running   0          2m59s   10.10.72.1      test-worker4   <none>           <none>
nginx-ds-bmv2l   1/1     Running   0          3m      10.10.244.69    test-worker8   <none>           <none>
nginx-ds-j47sp   1/1     Running   0          2m58s   10.10.8.129     test-worker5   <none>           <none>
nginx-ds-j7mlw   1/1     Running   0          3m      10.10.129.193   test-worker7   <none>           <none>
nginx-ds-jc9mj   1/1     Running   0          13s     10.10.4.2       test-worker2   <none>           <none>
nginx-ds-pqd27   1/1     Running   0          2m59s   10.10.222.65    test-worker    <none>           <none>
nginx-ds-rj24g   1/1     Running   0          2m58s   10.10.4.129     test-worker3   <none>           <none>
nginx-ds-zstmr   1/1     Running   0          3m      10.10.247.65    test-worker6   <none>           <none>
❯ kubectl exec nginx-ds-bmv2l -- curl -m 5 -s 10.10.4.2 > /dev/null
command terminated with exit code 28
```

## Multus-CNI

Multus CNI is a CNI plugin for Kubernetes that enables attaching multiple network interfaces to pods. It means that we
can use Multus CNI to attach both Calico and Antrea CNI to a Pod. In this way, we can migrate CNI from Calico to Antrea.

After deploying Calico and Antrea, we can use Multus CNI to attach both Calico and Antrea CNI to a Pod, which is a
Multus Pod.
The Multus Pods could communicate with both Calico Pods and Antrea Pods. After that, we can migrate CNI from Calico to
Antrea.

### Install Multus-CNI

```bash
# Install Multus-CNI
❯ kubectl apply -f https://raw.githubusercontent.com/intel/multus-cni/master/images/multus-daemonset.yml
daemonset.apps/multus created
# Change Multus-CNI configuration
# I use KinD to deploy Kubernetes, so I just run "docker exec"
# Move ./00-multus.conf to /etc/cni/net.d/00-multus.conf on all Nodes
nodes=$(kubectl get node --no-headers -o custom-columns=NAME:.metadata.name)
for node in $nodes; do
  echo "========== Move ./00-multus.conf to /etc/cni/net.d/00-multus.conf on node ${node} =========="
  docker cp ./00-multus.conf "${node}":/etc/cni/net.d/00-multus.conf
done
```

### Using Multus-CNI to be a bridge between Calico and Antrea

```bash
# A Cluster with Calico CNI
❯ kubectl get pod -o wide
NAME             READY   STATUS    RESTARTS   AGE     IP              NODE            NOMINATED NODE   READINESS GATES
nginx-ds-6d7bf   1/1     Running   0          5m3s    10.10.126.129   test-worker9    <none>           <none>
nginx-ds-9mnbn   1/1     Running   0          4m54s   10.10.39.193    test-worker2    <none>           <none>
nginx-ds-bnj62   1/1     Running   0          5m3s    10.10.4.129     test-worker3    <none>           <none>
nginx-ds-dfqpk   1/1     Running   0          5m4s    10.10.244.65    test-worker8    <none>           <none>
nginx-ds-m5jjp   1/1     Running   0          5m4s    10.10.247.65    test-worker6    <none>           <none>
nginx-ds-q9r7j   1/1     Running   0          5m3s    10.10.222.65    test-worker     <none>           <none>
nginx-ds-qtjs5   1/1     Running   0          5m4s    10.10.253.5     test-worker10   <none>           <none>
nginx-ds-rn4h7   1/1     Running   0          5m4s    10.10.72.1      test-worker4    <none>           <none>
nginx-ds-tbvpk   1/1     Running   0          4m59s   10.10.8.129     test-worker5    <none>           <none>
nginx-ds-tgxfs   1/1     Running   0          5m3s    10.10.129.193   test-worker7    <none>           <none>
# Install Antrea and Multus-CNI
❯ kubectl apply -f antrea.yml
❯ kubectl apply -f multus-daemonset.yml
# restart a random nginx Pod
❯ kubectl delete pod nginx-ds-qtjs5
pod "nginx-ds-qtjs5" deleted
```

After that, we can see that the nginx Pod is a Multus Pod with two network interfaces. One is managed by Calico CNI(
eth0), and
the other is managed by Antrea CNI(net1).

```bash
❯ kubectl exec nginx-ds-7m9jw -- ip a
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
    inet6 ::1/128 scope host
       valid_lft forever preferred_lft forever
2: tunl0@NONE: <NOARP> mtu 1480 qdisc noop state DOWN group default qlen 1000
    link/ipip 0.0.0.0 brd 0.0.0.0
4: eth0@if14: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1480 qdisc noqueue state UP group default qlen 1000
    link/ether c2:e7:5d:f9:81:03 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.10.253.7/32 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::c0e7:5dff:fef9:8103/64 scope link
       valid_lft forever preferred_lft forever
5: net1@if15: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UP group default
    link/ether 16:82:9b:e1:e4:75 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.10.1.2/24 brd 10.10.1.255 scope global net1
       valid_lft forever preferred_lft forever
    inet6 fe80::1482:9bff:fee1:e475/64 scope link
       valid_lft forever preferred_lft forever
```

The Multus Pods could still communicate with Calico Pods.

```bash
# Pod nginx-ds-7m9jw is a Multus Pod and Pod nginx-ds-tbvpk is a Calico Pod
❯ kubectl exec nginx-ds-tbvpk -- curl -m 5 10.10.253.7 > /dev/null
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   615  100   615    0     0    603      0  0:00:01  0:00:01 --:--:--   603
❯ kubectl exec nginx-ds-7m9jw -- curl -m 5 10.10.72.1 > /dev/null
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   615  100   615    0     0   494k      0 --:--:-- --:--:-- --:--:--  600k
```

If we remove the Multus CNI configuration file `00-multus.conf`, Kubelet will use Antrea CNI configuration file
`10-antrea.conflist`. So all Pods created after removing `00-multus.conf` will use Antrea CNI. They are Antrea Pods.

```bash
# Remove Multus CNI configuration file
nodes=$(kubectl get node --no-headers -o custom-columns=NAME:.metadata.name)
for node in $nodes; do
  echo "========== Remove Multus CNI configuration file on node ${node} =========="
  docker exec "${node}" rm /etc/cni/net.d/00-multus.conf
done
# Run a new Pod
❯ kubectl run antrea-nginx --image=nginx --image-pull-policy=Never
# The legacy Calico Pods could not communicate with the new Antrea Pods
# Pod nginx-ds-6d7bf is a Calico Pod.
❯ kubectl exec nginx-ds-6d7bf -- curl -m 5 10.10.5.2
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:--  0:00:05 --:--:--     0
curl: (28) Connection timed out after 5002 milliseconds
command terminated with exit code 28
# The legacy Multus Pod could communicate with the new Antrea Pods
# Pod nginx-ds-7m9jw is a Multus Pod.
❯ kubectl exec nginx-ds-7m9jw -- curl -m 5 10.10.5.2 > /dev/null
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   615  100   615    0     0   325k      0 --:--:-- --:--:-- --:--:--  600k
```

### PodCIDR

Does every Multus Pod could communicate with both Calico Pods and Antrea Pods? The answer is no. For example:

```bash
# Pod client is a Multus Pod and other Pods are Calico Pods
❯ kp -owide
NAME             READY   STATUS    RESTARTS   AGE     IP              NODE            NOMINATED NODE   READINESS GATES
client           1/1     Running   0          94s     10.10.20.2      test-worker2    <none>           <none>
nginx-ds-42xt6   1/1     Running   0          3m40s   10.10.115.1     test-worker17   <none>           <none>
nginx-ds-4dsp8   1/1     Running   0          3m42s   10.10.79.65     test-worker19   <none>           <none>
nginx-ds-57w84   1/1     Running   0          3m42s   10.10.93.1      test-worker15   <none>           <none>
nginx-ds-7qht2   1/1     Running   0          3m44s   10.10.244.65    test-worker8    <none>           <none>
nginx-ds-b7rh4   1/1     Running   0          3m43s   10.10.247.65    test-worker6    <none>           <none>
nginx-ds-d2nt5   1/1     Running   0          3m43s   10.10.6.65      test-worker11   <none>           <none>
nginx-ds-fg4jp   1/1     Running   0          3m43s   10.10.253.129   test-worker16   <none>           <none>
nginx-ds-fkrdd   1/1     Running   0          3m40s   10.10.39.193    test-worker2    <none>           <none>
nginx-ds-fxpkl   1/1     Running   0          3m44s   10.10.226.129   test-worker18   <none>           <none>
nginx-ds-lgkp7   1/1     Running   0          3m37s   10.10.129.193   test-worker7    <none>           <none>
nginx-ds-lrrln   1/1     Running   0          3m43s   10.10.253.1     test-worker10   <none>           <none>
nginx-ds-m68m7   1/1     Running   0          3m41s   10.10.222.65    test-worker     <none>           <none>
nginx-ds-n274g   1/1     Running   0          3m45s   10.10.134.131   test-worker13   <none>           <none>
nginx-ds-n29lb   1/1     Running   0          3m43s   10.10.8.129     test-worker5    <none>           <none>
nginx-ds-qhht2   1/1     Running   0          3m42s   10.10.85.1      test-worker14   <none>           <none>
nginx-ds-qmczc   1/1     Running   0          3m42s   10.10.123.129   test-worker20   <none>           <none>
nginx-ds-r2qjj   1/1     Running   0          3m43s   10.10.126.129   test-worker9    <none>           <none>
nginx-ds-rbr6n   1/1     Running   0          3m37s   10.10.4.129     test-worker3    <none>           <none>
nginx-ds-x7hbf   1/1     Running   0          3m42s   10.10.209.65    test-worker12   <none>           <none>
nginx-ds-xznvq   1/1     Running   0          3m45s   10.10.72.1      test-worker4    <none>           <none>

❯ kubectl exec client -- curl -m 5 -s 10.10.6.65
command terminated with exit code 28
❯ kubect exec client -- curl -m 5 -s 10.10.8.129
command terminated with exit code 28
❯ kubectl exec client -- curl -m 5 -s 10.10.4.129
command terminated with exit code 28
```

We notice that the Multus Pod client could not communicate with the Calico Pods on Node test-worker3, test-worker5, and
test-worker11.
The root cause is that the PodCIDR of these Nodes is `10.10.4.128/26`, `10.10.8.128/26`, and `10.10.6.42/26`, which is
conflicted with the PodCIDR of Antrea CNI(Node PodCIDR).
Which result in the conflict of the routing table.

```bash 
root@test-worker2:/# ip r
default via 172.18.0.1 dev eth0
10.10.0.0/24 via 10.10.0.1 dev antrea-gw0 onlink
10.10.1.0/24 via 10.10.1.1 dev antrea-gw0 onlink
10.10.2.0/24 via 10.10.2.1 dev antrea-gw0 onlink
10.10.3.0/24 via 10.10.3.1 dev antrea-gw0 onlink
10.10.4.0/24 via 10.10.4.1 dev antrea-gw0 onlink6
10.10.4.128/26 via 172.18.0.19 dev tunl0 proto bird onlink 
10.10.5.0/24 via 10.10.5.1 dev antrea-gw0 onlink
10.10.6.0/24 via 10.10.6.1 dev antrea-gw0 onlink
10.10.6.64/26 via 172.18.0.6 dev tunl0 proto bird onlink
10.10.7.0/24 via 10.10.7.1 dev antrea-gw0 onlink
10.10.8.0/24 via 10.10.8.1 dev antrea-gw0 onlink
10.10.8.128/26 via 172.18.0.15 dev tunl0 proto bird onlink
10.10.9.0/24 dev antrea-gw0 proto kernel scope link src 10.10.9.1
```

## Step to Migration CNI from Calico to Antrea

### Simple Scenario

If the PodCIDR of Calico ipamblock is not overlapped with the PodCIDR of Antrea(Node PodCIDR). We can migrate CNI from
Calico to Antrea by the following steps:

+ Print requirements of the migration.
+ Install Antrea and Multus-CNI with double CNI configuration files.
+ Convert Calico Networkpolicy to Antrea Networkpolicy.
+ Restart all Pods, then the Pods in cluster will be Multus Pods.
+ Remove Multus-CNI configuration file.
+ Restart all Pods, then the Pods in cluster will be Antrea Pods.
+ Remove Calico CNI.

For exmaple:

```bash
# Cluster with Calico CNI, and the PodCIDR of Calico ipamblock is not overlapped with the PodCIDR of Antrea(Node PodCIDR)
# Now both nginx-ds and client are Calico Pods
❯ kubectl get pod -o wide
NAME             READY   STATUS    RESTARTS   AGE     IP              NODE            NOMINATED NODE   READINESS GATES
client           1/1     Running   0          69s     10.10.253.2     test-worker10   <none>           <none>
nginx-ds-2p2vv   1/1     Running   0          3m44s   10.10.39.193    test-worker2    <none>           <none>
nginx-ds-2z4tk   1/1     Running   0          3m43s   10.10.123.129   test-worker20   <none>           <none>
nginx-ds-4p4ps   1/1     Running   0          3m46s   10.10.222.65    test-worker     <none>           <none>
nginx-ds-5dlsc   1/1     Running   0          3m44s   10.10.244.65    test-worker8    <none>           <none>
nginx-ds-5stzf   1/1     Running   0          3m43s   10.10.126.129   test-worker9    <none>           <none>
nginx-ds-6ddcc   1/1     Running   0          3m45s   10.10.253.129   test-worker16   <none>           <none>
nginx-ds-6wrpq   1/1     Running   0          3m43s   10.10.93.1      test-worker15   <none>           <none>
nginx-ds-7blzg   1/1     Running   0          3m44s   10.10.253.1     test-worker10   <none>           <none>
nginx-ds-dk8sc   1/1     Running   0          3m44s   10.10.134.129   test-worker13   <none>           <none>
nginx-ds-gdzwm   1/1     Running   0          3m44s   10.10.85.1      test-worker14   <none>           <none>
nginx-ds-hkv89   1/1     Running   0          3m45s   10.10.129.193   test-worker7    <none>           <none>
nginx-ds-ht6br   1/1     Running   0          3m43s   10.10.79.65     test-worker19   <none>           <none>
nginx-ds-p29g8   1/1     Running   0          3m43s   10.10.72.1      test-worker4    <none>           <none>
nginx-ds-rbs85   1/1     Running   0          3m46s   10.10.115.1     test-worker17   <none>           <none>
nginx-ds-tvklx   1/1     Running   0          3m43s   10.10.209.65    test-worker12   <none>           <none>
nginx-ds-w6hfg   1/1     Running   0          3m46s   10.10.247.65    test-worker6    <none>           <none>
nginx-ds-wnzml   1/1     Running   0          3m50s   10.10.226.130   test-worker18   <none>           <none>
❯ kubectl get svc
NAME           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
kubernetes     ClusterIP   10.96.0.1       <none>        443/TCP   9m16s
nginx-origin   ClusterIP   10.96.117.185   <none>        80/TCP    87s
```

Check the connectivity from client to Service, the Service `nginx-origin` works well. The traffic from client to Service
is from Calico Pod to Calico Pod.

```bash
❯ kubectl exec client -- /client -c 10 -n 1000 -url http://nginx-orign
2023/10/14 12:13:14 100 Requests completed
2023/10/14 12:13:16 200 Requests completed
2023/10/14 12:13:18 300 Requests completed
2023/10/14 12:13:20 400 Requests completed
2023/10/14 12:13:22 500 Requests completed
2023/10/14 12:13:24 600 Requests completed
2023/10/14 12:13:26 700 Requests completed
2023/10/14 12:13:28 800 Requests completed
2023/10/14 12:13:30 900 Requests completed
2023/10/14 12:13:32 1000 Requests completed
2023/10/14 12:13:32 Receiving Stop Signal, stopping...
2023/10/14 12:13:32 Receiving Stop Signal, stopping...
2023/10/14 12:13:32 Receiving Stop Signal, stopping...
2023/10/14 12:13:32 Receiving Stop Signal, stopping...
2023/10/14 12:13:32 Receiving Stop Signal, stopping...
2023/10/14 12:13:32 Receiving Stop Signal, stopping...
2023/10/14 12:13:32 Receiving Stop Signal, stopping...
2023/10/14 12:13:32 Receiving Stop Signal, stopping...
2023/10/14 12:13:32 Receiving Stop Signal, stopping...
2023/10/14 12:13:32 Receiving Stop Signal, stopping...
2023/10/14 12:13:32 Total Requests:  1002
2023/10/14 12:13:32 Success: 0
2023/10/14 12:13:32 Failure: 1002
2023/10/14 12:13:32 Success Rate: 0.000000%
2023/10/14 12:13:32 Total time: 24.651842152s
```

After deploying Antrea CNI and Multus, check the connectivity from client to Service again Service works well, too. The
traffic from client to Service is still from Calico Pod to Calico Pod because all Pods are legacy Calico Pods.

```bash
❯ kubectl exec client -- /client -c 10 -n 1000 -url http://nginx-orign
2023/10/14 12:15:01 100 Requests completed
2023/10/14 12:15:03 200 Requests completed
2023/10/14 12:15:05 300 Requests completed
2023/10/14 12:15:07 400 Requests completed
2023/10/14 12:15:09 500 Requests completed
2023/10/14 12:15:11 600 Requests completed
2023/10/14 12:15:13 700 Requests completed
2023/10/14 12:15:15 800 Requests completed
2023/10/14 12:15:17 900 Requests completed
2023/10/14 12:15:19 1000 Requests completed
2023/10/14 12:15:19 Receiving Stop Signal, stopping...
2023/10/14 12:15:19 Receiving Stop Signal, stopping...
2023/10/14 12:15:19 Receiving Stop Signal, stopping...
2023/10/14 12:15:19 Receiving Stop Signal, stopping...
2023/10/14 12:15:19 Receiving Stop Signal, stopping...
2023/10/14 12:15:19 Receiving Stop Signal, stopping...
2023/10/14 12:15:19 Receiving Stop Signal, stopping...
2023/10/14 12:15:19 Receiving Stop Signal, stopping...
2023/10/14 12:15:19 Receiving Stop Signal, stopping...
2023/10/14 12:15:19 Receiving Stop Signal, stopping...
2023/10/14 12:15:19 Total Requests:  1003
2023/10/14 12:15:19 Success: 0
2023/10/14 12:15:19 Failure: 1003
2023/10/14 12:15:19 Success Rate: 0.000000%
2023/10/14 12:15:19 Total time: 20.259637559s
```

Create a new client `multus-client` with Multus CNI, Check the connectivity from multus-client to Calico Nginx. The
traffic from multus-client to Service is from Multus Pod to Calico Pod.

```bash
❯ kubectl exec multus-client -- /client -c 10 -n 1000 -url http://nginx-origin
2023/10/14 12:23:32 100 Requests completed
2023/10/14 12:23:34 200 Requests completed
2023/10/14 12:23:36 300 Requests completed
2023/10/14 12:23:38 400 Requests completed
2023/10/14 12:23:40 500 Requests completed
2023/10/14 12:23:42 600 Requests completed
2023/10/14 12:23:44 700 Requests completed
2023/10/14 12:23:46 800 Requests completed
2023/10/14 12:23:48 900 Requests completed
2023/10/14 12:23:50 1000 Requests completed
2023/10/14 12:23:50 Receiving Stop Signal, stopping...
2023/10/14 12:23:50 Receiving Stop Signal, stopping...
2023/10/14 12:23:50 Receiving Stop Signal, stopping...
2023/10/14 12:23:50 Receiving Stop Signal, stopping...
2023/10/14 12:23:50 Receiving Stop Signal, stopping...
2023/10/14 12:23:50 Receiving Stop Signal, stopping...
2023/10/14 12:23:50 Receiving Stop Signal, stopping...
2023/10/14 12:23:50 Receiving Stop Signal, stopping...
2023/10/14 12:23:50 Receiving Stop Signal, stopping...
2023/10/14 12:23:50 Receiving Stop Signal, stopping...
2023/10/14 12:23:50 Total Requests:  1005
2023/10/14 12:23:50 Success: 1005
2023/10/14 12:23:50 Failure: 0
2023/10/14 12:23:50 Success Rate: 100.000000%
2023/10/14 12:23:50 Total time: 20.414199716s
```

Restart nginx daemonset, after restart. Remove Multus CNI. After this, the priority of Antrea CNI is higher than Calico,
so all new Pods will be Antrea Pods.
Check the connectivity from Antrea client to Multus Nginx, it still works well.

```bash
❯ kubectl exec antrea-client -- /client -c 10 -n 1000 -url http://nginx-origin
2023/10/14 12:27:32 100 Requests completed
2023/10/14 12:27:35 200 Requests completed
2023/10/14 12:27:37 300 Requests completed
2023/10/14 12:27:39 400 Requests completed
2023/10/14 12:27:41 500 Requests completed
2023/10/14 12:27:43 600 Requests completed
2023/10/14 12:27:45 700 Requests completed
2023/10/14 12:27:47 800 Requests completed
2023/10/14 12:27:51 900 Requests completed
2023/10/14 12:27:53 1000 Requests completed
2023/10/14 12:27:53 Receiving Stop Signal, stopping...
2023/10/14 12:27:53 Receiving Stop Signal, stopping...
2023/10/14 12:27:53 Receiving Stop Signal, stopping...
2023/10/14 12:27:53 Receiving Stop Signal, stopping...
2023/10/14 12:27:53 Receiving Stop Signal, stopping...
2023/10/14 12:27:53 Receiving Stop Signal, stopping...
2023/10/14 12:27:53 Receiving Stop Signal, stopping...
2023/10/14 12:27:53 Receiving Stop Signal, stopping...
2023/10/14 12:27:53 Receiving Stop Signal, stopping...
2023/10/14 12:27:53 Receiving Stop Signal, stopping...
2023/10/14 12:27:53 Total Requests:  1008
2023/10/14 12:27:53 Success: 1008
2023/10/14 12:27:53 Failure: 0
2023/10/14 12:27:53 Success Rate: 100.000000%
2023/10/14 12:27:53 Total time: 24.287786997s
```

Restart nginx again, then all Pods will only use Antrea CNI with single NIC. Check the connectivity from Antrea client to Nginx.
It still works well. The traffic from Antrea client to Nginx is from Antrea Pod to Antrea Pod.

```bash
❯ k exec antrea-client -- /client -c 10 -n 1000 -url http://nginx-origin
2023/10/14 12:35:44 100 Requests completed
2023/10/14 12:35:46 200 Requests completed
2023/10/14 12:35:48 300 Requests completed
2023/10/14 12:35:50 400 Requests completed
2023/10/14 12:35:52 500 Requests completed
2023/10/14 12:35:54 600 Requests completed
2023/10/14 12:35:56 700 Requests completed
2023/10/14 12:35:58 800 Requests completed
2023/10/14 12:36:00 900 Requests completed
2023/10/14 12:36:02 1000 Requests completed
2023/10/14 12:36:02 Receiving Stop Signal, stopping...
2023/10/14 12:36:02 Receiving Stop Signal, stopping...
2023/10/14 12:36:02 Receiving Stop Signal, stopping...
2023/10/14 12:36:02 Receiving Stop Signal, stopping...
2023/10/14 12:36:02 Receiving Stop Signal, stopping...
2023/10/14 12:36:02 Receiving Stop Signal, stopping...
2023/10/14 12:36:02 Receiving Stop Signal, stopping...
2023/10/14 12:36:02 Receiving Stop Signal, stopping...
2023/10/14 12:36:02 Receiving Stop Signal, stopping...
2023/10/14 12:36:02 Receiving Stop Signal, stopping...
2023/10/14 12:36:02 Total Requests:  1003
2023/10/14 12:36:02 Success: 1003
2023/10/14 12:36:02 Failure: 0
2023/10/14 12:36:02 Success Rate: 100.000000%
2023/10/14 12:36:02 Total time: 20.248457447s
```

### A Small Part of PodCIDR Overlapped

+ Print requirements of the migration.
+ Drain the Node with overlapped PodCIDR.
+ Install Antrea and Multus-CNI with double CNI configuration files.
+ Convert Calico Networkpolicy to Antrea Networkpolicy.
+ Restart all Pods, then the Pods in cluster will be Multus Pods.
+ Remove Multus-CNI configuration file.
+ Restart all Pods, then the Pods in cluster will be Antrea Pods.
+ Remove Calico CNI.
+ Uncordon the Node with overlapped PodCIDR.

### Most of PodCIDR Overlapped

#### Solution 1

Edit the Calico IPPool to avoid PodCIDR overlapped.

#### Solution 2

Hard way. We need to migrate CNI from Calico to Antrea in a small part of Nodes. Then we need to migrate CNI from Calico
to Antrea in the rest of Nodes.
The connection between the two parts of Nodes may be broken.

## Files to deliver

+ `antrea-migrate.sh`: A script to migrate CNI from Calico to Antrea.
+ `antrea-migrator`: A tool to convert Calico Networkpolicy to Antrea Networkpolicy, check CIDR overlapped, and so on.
+ `antrea-migrator` source code (Maybe we could add it to `antctl`)

## Reference
multus-cni: https://github.com/k8snetworkplumbingwg/multus-cni
some scripts: https://github.com/hjiajing/studious-doodle (WIP)

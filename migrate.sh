#!/bin/bash

set -e

echo "========== Begin to migrate Calico CNI to Antrea =========="
echo "========== Warning: Your Service in the Cluster may be not available while migration =========="
echo "========== Building Antrea-migrator binary execute file =========="
go build -o antrea-migrator ../main.go

# Creating Sandbox killer job on every Node
function create_kill_sandbox_job() {
    node=$1
    echo "========== Creating Sandbox killer job on Node: $node =========="
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
          image: harbor-repo.vmware.com/dockerhub-proxy-cache/busybox:1.28
          command: 
          - "/bin/sh"
          - "-c"
          - "pkill -9 /pause"
          securityContext:
            privileged: true
        restartPolicy: Never
EOF
}

function wait_for_kill_sandbox_job() {
    node=$1
    echo "========== Waiting for Sandbox killer job on Node: $node =========="
    kubectl wait --for=condition=complete --timeout=300s job/antrea-migrator-kill-sandbox-$node -n kube-system
}

function print_requirements() {
    ./antrea-migrator print-requirements
}

function remove_calico() {
    echo "========== Removing Calico =========="
    kubectl delete -f https://docs.projectcalico.org/manifests/calico.yaml
    echo "========== Waiting for Calico to be removed =========="
    kubectl wait --for=delete pod -l k8s-app=calico-node -n kube-system
    echo "========== Calico is removed =========="
}

function remove_migrate_job() {
    node=$1
    echo "========== Removing Sandbox killer job on Node: $node =========="
    kubectl delete job/antrea-migrator-kill-sandbox-$node -n kube-system
}

function remove_calico_iptables() {
    antrea_agents=$(kubectl get pods -n kube-system -l app=antrea -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')
    for antrea_agent in $antrea_agents; do
        echo "========== Removing Calico iptables rules on Node: $antrea_agent =========="
        kubectl exec  $antrea_agent -n kube-system -- /bin/sh -c "iptables-save | grep -v cali | iptables-restore"
    done
}

# function start_migrate is used to migrate Calico CNI to Antrea
function start_migrate() {
    echo "========== Deploying Antrea =========="
    kubectl apply -f https://raw.githubusercontent.com/antrea-io/antrea/master/build/yamls/antrea.yml
    echo "========== Waiting for Antrea to be ready =========="
    kubectl rollout status ds/antrea-agent -n kube-system
    ./antrea-migrator convert-networkpolicy
    nodes=$(kubectl get node -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')
    for node in $nodes; do
        create_kill_sandbox_job $node
    done
    for node in $nodes; do
        wait_for_kill_sandbox_job $node
    done
    echo "========== Antrea is ready =========="
    remove_calico
    for node in $nodes; do
        remove_migrate_job $node
    done
    remove_calico_iptables
}

start_migrate

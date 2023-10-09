#!/bin/bash

set -e

intersect_nodes_array=()

function coexist-migration() {
  echo "========== Step 2: Begin coexist migration =========="
  for node in $intersect_nodes_array; do
    echo "========== Taint node ${node} NoSchedule and NoExecute =========="
    kubectl taint node "${node}" antrea-migration=:NoSchedule
    kubectl taint node "${node}" antrea-migration=:NoExecute
  done
  echo "========== Deploying Antrea agent =========="
  kubectl apply -f antrea.yml
  kubectl rollout status -n kube-system ds antrea-agent
  echo "========== Converting Network Policies =========="
  ./antrea-migrator convert-networkpolicy
}

function no-coexist-migration() {
  echo "========== In this step, CNI will be replaced by Antrea =========="
  echo $intersect_nodes_array
  for node in $intersect_nodes_array; do
    echo "========== Taint node ${node} NoSchedule and NoExecute =========="
    kubectl taint node "${node}" antrea-migration=:NoSchedule
    kubectl taint node "${node}" antrea-migration=:NoExecute
  done
}

function get-pod-cidr() {
  cluster_cidr=$(kubectl cluster-info dump | grep -oP '(?<=--cluster-cidr=)[^ ]*'  | sed 's/\"//g' | sed 's/,//g')
}

echo "========== Begin to migrate Calico CNI to Antrea =========="
echo "========== Warning: Your Service in the Cluster may be not available while migration =========="
echo "========== Building Antrea-migrator binary execute file =========="
go build -o antrea-migrator ../main.go

echo "========== Step 1: Checking Node CIDR =========="
intersect_nodes=$(./antrea-migrator check-nodes | tr -d "[]")
intersect_nodes_array=($intersect_nodes)
echo $intersect_nodes_array
node_number=$(kubectl get node --no-headers -o custom-columns=NAME:.metadata.name | wc -l)
quarter=$(expr "$node_number" / 4)
if [ "${#intersect_nodes_array[@]}" -gt "$quarter" ]; then
  echo "========== Begin to migrate Calico CNI to Antrea with no-coexist mode =========="
  no-coexist-migration
else
  echo "========== Begin to migrate Calico CNI to Antrea with coexist mode =========="
  coexist-migration
fi

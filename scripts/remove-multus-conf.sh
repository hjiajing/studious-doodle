#!/bin/bash


# remove /etc/cni/net.d/00-multus.conf on all Nodes
nodes=$(kubectl get node --no-headers -o custom-columns=NAME:.metadata.name)
for node in $nodes; do
  echo "========== Remove /etc/cni/net.d/00-multus.conf on node ${node} =========="
  docker exec -it "${node}" bash -c "rm -f /etc/cni/net.d/00-multus.conf"
  docker exec -it "${node}" bash -c "rm -rf /etc/cni/net.d/multus.d"
done
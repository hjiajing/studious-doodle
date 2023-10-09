#!/bin/bash

# Move ./00-multus.conf to /etc/cni/net.d/00-multus.conf on all Nodes
nodes=$(kubectl get node --no-headers -o custom-columns=NAME:.metadata.name)
for node in $nodes; do
  echo "========== Move ./00-multus.conf to /etc/cni/net.d/00-multus.conf on node ${node} =========="
  docker cp ./00-multus.conf "${node}":/etc/cni/net.d/00-multus.conf
done
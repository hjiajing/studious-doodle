#!/bin/bash

# Check connectivity from all Pods to all Pods on other Node
pods=$(kubectl get pod -n default --no-headers -o custom-columns=NAME:.metadata.name)
for pod in $pods; do
  ./check-connectivity-one-to-all.sh "${pod}" default
done
#!/bin/bash

# check connectivity from one Pod to all Pods on other Node
# usage: ./check-connectivity-one-to-all.sh <pod-name> <namespace>
# e.g. ./check-connectivity-one-to-all.sh busybox1 default
pod=$1
namespace=$2
# get pods IP
podip=$(kubectl get pod -n "${namespace}" -o wide --no-headers | awk '{print $6}')
for ip in $podip; do
  kubectl exec "${pod}" -n "${namespace}" -- curl -s -m 5 "${ip}":80 > /dev/null
  if [ $? -eq 0 ]; then
    echo "Pod ${pod} can connect to ${ip}:80"
  else
    echo "Pod ${pod} cannot connect to ${ip}:80"
  fi
done
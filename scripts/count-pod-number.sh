#!/bin/bash

# count the number of Pods in the Cluster
pod_number=$(kubectl get pod --all-namespaces --no-headers | wc -l)

# count the IP number of IP Pool
cidr=$(kubectl get ippool default-ipv4-ippool -o jsonpath='{.spec.cidr}')
echo $cidr
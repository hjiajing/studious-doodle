#/bin/bash

dest=$1
pods=$(kubectl get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}')
for pod in $pods; do
  echo "========== Checking connectivity from $pod to $dest =========="
  kubectl exec -it $pod -- curl -s -m 5 $dest > /dev/null
done
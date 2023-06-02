#!/usr/bin/env bash
set -u

# sample script to get a log of all the mutations for all the pods for a particular ROKT PSP e.g. restricted
# cat output to grep "is mutated by" | cut -d : -f2- | sort |  uniq to get a view of all distinct mutations
#
export LEVEL="${1}"

NAMESPACES=$(kubectl get NAMESPACES --no-headers -o custom-columns=":metadata.name")

for NAMESPACE in $NAMESPACES
do
  echo "NAMESPACE=$NAMESPACE"
  if kubectl get rolebinding --namespace "$NAMESPACE" -o name | grep "$LEVEL-psp-rolebinding" > /dev/null
  then
    PODS=$(kubectl get pods -n "$NAMESPACE" --no-headers -o custom-columns=":metadata.name")
    for POD in $PODS
    do
      echo "POD=$POD"
      echo "$NAMESPACE"/"$POD"
      if ! [[ "$POD" == *"istio-ingressgateway"* ]]
      then
        # exclude the sidecars we know of to reduce the diff output
        pspmigrator mutating pod -n "$NAMESPACE" "$POD" -c flipt-daemon,istio-validation,istio-proxy
      else
        pspmigrator mutating pod -n "$NAMESPACE" "$POD"
      fi
    done
  fi
done

#!/bin/bash

if [[ $1 == "--config" ]] ; then
  cat <<EOF
configVersion: v1
kubernetes:
- apiVersion: v1
  kind: Pod
  executeHookOnEvent: ["Added", "Deleted","Modified"]
  labelSelector:
    matchLabels:
      app: registry
  namespace:
    nameSelector:
      matchNames: ["default"]
EOF
else
  type=$(jq -r '.[0].type' $BINDING_CONTEXT_PATH)
  if [[ $type == "Event" ]] ; then
    podName=$(jq -r '.[0].object.metadata.name' $BINDING_CONTEXT_PATH)
    echo "Pod '${podName}' added"
    rates=$(jq -r '.[0].object.metadata.annotations.egress-bandwidth' $BINDING_CONTEXT_PATH)
    port=$(jq -r '.[0].object.spec.containers.')
    container=$()
    hostnetwork=()
    #set qdisc rules for host
  fi
fi

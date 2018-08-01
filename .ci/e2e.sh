#!/bin/bash

set -exuo pipefail

cd $(dirname $0)/..

kubectl apply -f ./examples/deployment.yaml

set +e
while true
do
    kubectl get cm to-sync -n default -o json | jq -re .data.foo && break
    sleep 5
    echo "====="
    kubectl logs -l app=kube-sync -n kube-system
    echo "====="
done

kubectl apply -f ./.ci/configmap-v2.yaml

while true
do
    kubectl get cm to-sync -n default -o json | jq -re .data.bar && break
    sleep 5
    echo "====="
    kubectl logs -l app=kube-sync -n kube-system
    echo "====="
done

NS=ns-$(uuidgen)
kubectl create ns ${NS}

while true
do
    kubectl get cm to-sync -n ${NS} -o json | jq -re .data.bar && break
    sleep 5
    echo "====="
    kubectl logs -l app=kube-sync -n kube-system
    echo "====="
done

set -e
for ns in $(kubectl get ns -o json | jq -re .items[].metadata.name)
do
    kubectl get cm to-sync -o yaml -n ${ns}
done

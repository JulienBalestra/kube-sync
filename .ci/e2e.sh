#!/bin/bash

set -exuo pipefail

cd $(dirname $0)/..

kubectl apply -f .ci/argo/controller/

argo version || {
    curl -Lf https://github.com/argoproj/argo/releases/download/v2.1.1/argo-linux-amd64 -o ${GOPATH}/bin/argo
    chmod +x ${GOPATH}/bin/argo
}

argo submit .ci/workflow.yaml
argo logs -w kube-sync -f


#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly KIND_VERSION=0.1.0
readonly CLUSTER_NAME=chart-testing
readonly K8S_VERSION=v1.13.2

create_kind_cluster() {
    kind create cluster --name "$CLUSTER_NAME" --config examples/kind/test/kind-config.yaml --image "kindest/node:$K8S_VERSION" --wait 10s
    KUBECONFIG="$(kind get kubeconfig-path --name=$CLUSTER_NAME)"
    export KUBECONFIG

    kubectl cluster-info || kubectl cluster-info dump
    echo

    echo -n 'Waiting for cluster to be ready...'
    until ! grep --quiet 'NotReady' <(kubectl get nodes --no-headers); do
        printf '.'
        sleep 1
    done

    echo '✔︎'
    echo

    kubectl get nodes
    echo

    echo 'Cluster ready!'
    echo
}

install_tiller() {
    echo 'Installing Tiller...'
    kubectl --namespace kube-system --output yaml create serviceaccount tiller --dry-run | kubectl apply -f -
    kubectl create --output yaml clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller --dry-run | kubectl apply -f -
    helm init --service-account tiller --upgrade --wait
    echo
}

install_local-path-provisioner() {
    # kind doesn't support Dynamic PVC provisioning yet, this is one way to get it working
    # https://github.com/rancher/local-path-provisioner

    # Remove default storage class. It will be recreated by local-path-provisioner
    kubectl delete storageclass standard

    echo 'Installing local-path-provisioner...'
    kubectl apply -f examples/kind/test/local-path-provisioner.yaml
    echo
}

test_e2e() {
    go test -cover -race -tags=integration ./...
    echo
}

cleanup() {
    kind delete cluster --name "$CLUSTER_NAME"
    echo 'Done!'
}

main() {
    trap cleanup EXIT

    create_kind_cluster
    install_local-path-provisioner
    install_tiller
    test_e2e
}

main

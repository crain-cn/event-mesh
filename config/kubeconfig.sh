#!/usr/bin/env bash

serviceaccount="eventmesh"
namespace="jituan-zhongtai-iaas"

secret=$(kubectl get secret -n $namespace  | grep $serviceaccount |awk '{print $1}')
echo $secret

# token
ca_crt_data="$(kubectl get secret "$secret" -n "$namespace" -o "jsonpath={.data.ca\.crt}" | openssl enc -d -base64 -A)"
token="$(kubectl get secret "$secret" -n "$namespace" -o "jsonpath={.data.token}" | openssl enc -d -base64 -A)"

# context
context="$(kubectl config current-context)"
# cluster
cluster="$(kubectl config view -o "jsonpath={.contexts[?(@.name==\"$context\")].context.cluster}")"
server="$(kubectl config view -o "jsonpath={.clusters[?(@.name==\"$cluster\")].cluster.server}")"

rm -rf ./kube.config && touch ./kube.config

export KUBECONFIG="./kube.config"
kubectl config set-credentials "$serviceaccount" --token="$token" >/dev/null
ca_crt="$(mktemp)"; echo "$ca_crt_data" > $ca_crt
kubectl config set-cluster "$cluster" --server="$server" --certificate-authority="$ca_crt" --embed-certs >/dev/null
kubectl config set-context "$context" --cluster="$cluster" --namespace="$namespace" --user="$serviceaccount" >/dev/null
kubectl config use-context "$context" >/dev/null

cat "$KUBECONFIG"

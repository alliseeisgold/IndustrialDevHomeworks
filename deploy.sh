#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TASK_DIR="$ROOT_DIR/task1"
K8S_DIR="$TASK_DIR/k8s"

log() {
  printf '[deploy] %s\n' "$1"
}

wait_for_service_endpoints() {
  local retries=60
  local delay=2
  local endpoints=""

  for _ in $(seq 1 "$retries"); do
    endpoints="$(kubectl get endpoints custom-app-service -o jsonpath='{.subsets[*].addresses[*].ip}' 2>/dev/null || true)"
    if [ -n "$endpoints" ]; then
      return 0
    fi
    sleep "$delay"
  done

  log "service custom-app-service has no ready endpoints"
  return 1
}

if ! command -v kubectl >/dev/null 2>&1; then
  log "kubectl is required"
  exit 1
fi

if ! kubectl cluster-info >/dev/null 2>&1; then
  log "kubernetes cluster is not available"
  exit 1
fi

CURRENT_CONTEXT="$(kubectl config current-context 2>/dev/null || true)"

log "building custom-app image"
if [ "$CURRENT_CONTEXT" = "minikube" ] && command -v minikube >/dev/null 2>&1; then
  minikube image build -t custom-app:latest "$TASK_DIR"
else
  if ! command -v docker >/dev/null 2>&1; then
    log "docker is required when current context is not minikube"
    exit 1
  fi

  docker build -t custom-app:latest "$TASK_DIR"

  if command -v minikube >/dev/null 2>&1 && minikube status >/dev/null 2>&1; then
    minikube image load custom-app:latest
  fi
fi

log "applying configmap and standalone pod"
kubectl apply -f "$K8S_DIR/configmap.yaml"
kubectl apply -f "$K8S_DIR/pod.yaml"

log "applying deployment and service"
kubectl apply -f "$K8S_DIR/deployment.yaml"
kubectl apply -f "$K8S_DIR/service.yaml"

log "applying daemonset and cronjob resources"
kubectl apply -f "$K8S_DIR/log-agent-rbac.yaml"
kubectl apply -f "$K8S_DIR/daemonset.yaml"
kubectl apply -f "$K8S_DIR/cronjob.yaml"

if [ -f "$K8S_DIR/statefulset.yaml" ]; then
  log "applying statefulset"
  kubectl apply -f "$K8S_DIR/statefulset.yaml"
fi

log "waiting for standalone pod"
kubectl wait --for=condition=Ready pod/custom-app-pod --timeout=180s

log "waiting for deployment rollout"
kubectl rollout status deployment/custom-app-deployment --timeout=180s

log "waiting for daemonset rollout"
kubectl rollout status daemonset/log-agent --timeout=180s

log "waiting for service endpoints"
wait_for_service_endpoints

log "current resources"
kubectl get pod custom-app-pod
kubectl get deployment custom-app-deployment
kubectl get pods -l app=custom-app,component=deployment
kubectl get service custom-app-service
kubectl get daemonset log-agent
kubectl get pods -l app=log-agent
kubectl get cronjob log-archive

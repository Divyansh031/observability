#!/usr/bin/env bash
# Recreates the local kind cluster, deploys the app services, and installs
# the kube-prometheus-stack monitoring stack (Prometheus, Grafana,
# Alertmanager). We recreate this every session (cluster gets deleted at
# session end), so this script exists to make that repeatable and
# typo-proof rather than re-running a dozen manual commands each time.
set -euo pipefail

CLUSTER=observability
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

echo "==> Creating kind cluster '$CLUSTER'"
kind create cluster --name "$CLUSTER"

echo "==> Building service images"
docker build -t inventory-service:dev "$REPO_ROOT/services/inventory-service"
docker build -t payment-service:dev "$REPO_ROOT/services/payment-service"
docker build -t order-service:dev "$REPO_ROOT/services/order-service"

echo "==> Loading images into kind"
kind load docker-image inventory-service:dev payment-service:dev order-service:dev --name "$CLUSTER"

echo "==> Applying app manifests"
kubectl apply -f "$SCRIPT_DIR"

echo "==> Waiting for app pods to be ready"
kubectl -n observability wait --for=condition=Ready pod --all --timeout=120s

echo "==> Installing kube-prometheus-stack (Prometheus + Grafana + Alertmanager)"
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts >/dev/null 2>&1 || true
helm repo update >/dev/null
helm install kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace \
  --wait --timeout 5m

echo "==> Applying ServiceMonitors and dashboards"
kubectl apply -k "$SCRIPT_DIR/monitoring/"

echo "==> Done. Cluster '$CLUSTER' is up, app + monitoring stack ready."

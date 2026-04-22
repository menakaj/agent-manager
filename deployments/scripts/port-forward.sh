#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/env.sh"

echo "=== Setting up Port Forwarding for OpenChoreo Services ==="
echo ""
echo "📋 Note: AMP services (Console, API, Traces Observer) use ClusterIP."
echo "   In local dev mode, they are accessed via port-forward below."
echo "   In Helm-based deployments, the AMP Gateway handles routing."
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed"
    exit 1
fi

# Check if cluster is running
if ! kubectl cluster-info --context $CLUSTER_CONTEXT &> /dev/null; then
    echo "❌ k3d cluster '$CLUSTER_NAME' is not running"
    exit 1
fi

echo "🔧 Setting kubectl context..."
kubectl config use-context $CLUSTER_CONTEXT

echo ""
echo "🌐 Starting port forwarding for OpenChoreo services..."
echo "   Press Ctrl+C to stop all port forwarding"
echo ""

# Function to cleanup background processes on exit
cleanup() {
    echo ""
    echo "🛑 Stopping all port forwarding..."
    jobs -p | xargs kill 2>/dev/null || true
    exit 0
}
trap cleanup EXIT INT TERM

# --- AMP Services (ClusterIP — require port-forward for local dev) ---

# Port forward Traces Observer Service
echo "🔍 Forwarding Traces Observer Service (9098)..."
kubectl port-forward -n openchoreo-observability-plane svc/amp-traces-observer 9098:9098 &

# Port forward Thunder (IDP)
echo "🔑 Forwarding Thunder IDP Service (8090)..."
kubectl port-forward -n amp-thunder svc/amp-thunder-extension-service 8090:8090 &

# Port forward AI Gateway Runtime (ClusterIP — needs port-forward for local agents)
echo "🤖 Forwarding AI Gateway Runtime (8084)..."
kubectl port-forward -n openchoreo-data-plane svc/default-ai-gateway-gateway-runtime 8084:8084 &

# --- Observability Services ---

# Port forward Observability Gateway (OTel ingestion)
echo "🌐 Forwarding Observability Gateway HTTP (22893)..."
kubectl port-forward -n openchoreo-data-plane svc/obs-gateway-gateway-gateway-runtime 22893:22893 &

echo "🌐 Forwarding Observability Gateway HTTPS (22894)..."
kubectl port-forward -n openchoreo-data-plane svc/obs-gateway-gateway-gateway-runtime 22894:22894 &

# --- Internal / Debug Services ---

# Port forward OpenSearch (for debugging)
echo "📊 Forwarding OpenSearch (9200)..."
kubectl port-forward -n openchoreo-observability-plane svc/opensearch 9200:9200 &

# Port forward Observer Service API (internal)
echo "🔍 Forwarding Observer Service API (8085)..."
kubectl port-forward -n openchoreo-observability-plane svc/observer 8085:8080 &

# Port forward OpenBao (Data Plane Secrets)
echo "🔐 Forwarding OpenBao Data Plane (8200)..."
kubectl port-forward -n amp-secrets svc/amp-secrets-openbao 8200:8200 &

# Port forward OpenBao (Workflow Plane - Git Secrets)
echo "🔐 Forwarding OpenBao Workflow Plane (8201)..."
kubectl port-forward -n openbao svc/openbao 8201:8200 &

# Port forward OpenChoreo API (internal only — not publicly exposed)
echo "🔧 Forwarding OpenChoreo API (8195)..."
kubectl port-forward svc/openchoreo-api -n openchoreo-control-plane 8195:8080 &

echo ""
echo "✅ Port forwarding active:"
echo ""
echo "   AMP Services (ClusterIP):"
echo "   Traces Observer Service:          http://localhost:9098"
echo "   Thunder IDP Service:              http://localhost:8090"
echo "   AI Gateway Runtime:               http://localhost:8084"
echo ""
echo "   Observability:"
echo "   Observability Gateway:            http://localhost:22893/otel"
echo "   Observability Gateway (HTTPS):    https://localhost:22894/otel"
echo ""
echo "   Internal / Debug:"
echo "   OpenSearch:                       http://localhost:9200"
echo "   Observer Service API:             http://localhost:8085"
echo "   OpenBao (Data Plane):             http://localhost:8200"
echo "   OpenBao (Workflow Plane):         http://localhost:8201"
echo "   OpenChoreo API (internal):        http://localhost:8195"

echo ""
echo "💡 Keep this terminal open. Press Ctrl+C to stop."

# Wait for all background jobs
wait

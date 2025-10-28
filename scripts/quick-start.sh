#!/bin/bash

# Quick start script for TimVault Operator
# This script helps you quickly deploy the operator to your Kubernetes cluster

set -e

echo "======================================"
echo "TimVault Operator - Quick Start"
echo "======================================"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if connected to a cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Not connected to a Kubernetes cluster. Please configure kubectl."
    exit 1
fi

echo "✅ Connected to Kubernetes cluster"
echo ""

# Step 1: Install CRDs
echo "📦 Installing Custom Resource Definitions..."
kubectl apply -f config/crd/timsecret-crd.yaml
kubectl apply -f config/crd/timsecretconfig-crd.yaml
echo "✅ CRDs installed"
echo ""

# Step 2: Create namespace
echo "📦 Creating namespace..."
kubectl apply -f config/manager/namespace.yaml
echo "✅ Namespace created"
echo ""

# Step 3: Install RBAC
echo "📦 Installing RBAC resources..."
kubectl apply -f config/rbac/service_account.yaml
kubectl apply -f config/rbac/role.yaml
kubectl apply -f config/rbac/role_binding.yaml
echo "✅ RBAC resources installed"
echo ""

# Step 4: Check if Docker image exists
IMAGE_NAME="timvault-operator:latest"
echo "⚠️  Note: Make sure you have built the Docker image: $IMAGE_NAME"
echo "   Run: make docker-build"
echo ""

read -p "Have you built the Docker image? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Please build the image first with: make docker-build"
    exit 1
fi

# Step 5: Deploy operator
echo "📦 Deploying operator..."
kubectl apply -f config/manager/deployment.yaml
echo "✅ Operator deployed"
echo ""

# Step 6: Wait for operator to be ready
echo "⏳ Waiting for operator to be ready..."
kubectl wait --for=condition=available --timeout=120s \
    deployment/timvault-operator-controller \
    -n timvault-operator-system

echo ""
echo "======================================"
echo "✅ Installation Complete!"
echo "======================================"
echo ""
echo "Next steps:"
echo "1. Create a TimSecret resource (see examples/timsecret-example.yaml)"
echo "2. Check the operator logs:"
echo "   kubectl logs -n timvault-operator-system deployment/timvault-operator-controller -f"
echo ""
echo "For more information, see README.md"


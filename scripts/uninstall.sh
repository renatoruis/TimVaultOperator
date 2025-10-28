#!/bin/bash

# Uninstall script for TimVault Operator

set -e

echo "======================================"
echo "TimVault Operator - Uninstall"
echo "======================================"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed."
    exit 1
fi

echo "⚠️  This will remove the TimVault Operator from your cluster."
read -p "Are you sure you want to continue? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Uninstall cancelled."
    exit 0
fi

# Step 1: Delete all TimSecret resources
echo "🗑️  Deleting TimSecret resources..."
kubectl delete timsecrets --all --all-namespaces || true
echo "✅ TimSecret resources deleted"
echo ""

# Step 2: Delete operator deployment
echo "🗑️  Deleting operator deployment..."
kubectl delete -f config/manager/deployment.yaml || true
echo "✅ Operator deployment deleted"
echo ""

# Step 3: Delete RBAC resources
echo "🗑️  Deleting RBAC resources..."
kubectl delete -f config/rbac/role_binding.yaml || true
kubectl delete -f config/rbac/role.yaml || true
kubectl delete -f config/rbac/service_account.yaml || true
echo "✅ RBAC resources deleted"
echo ""

# Step 4: Delete CRDs
echo "🗑️  Deleting Custom Resource Definitions..."
kubectl delete -f config/crd/timsecret-crd.yaml || true
kubectl delete -f config/crd/timsecretconfig-crd.yaml || true
echo "✅ CRDs deleted"
echo ""

# Step 5: Delete namespace
echo "🗑️  Deleting namespace..."
kubectl delete namespace timvault-operator-system || true
echo "✅ Namespace deleted"
echo ""

echo "======================================"
echo "✅ Uninstall Complete!"
echo "======================================"


#!/bin/bash

# Script to setup centralized Vault configuration using TimSecretConfig
# This creates a TimSecretConfig that can be used by all TimSecrets

set -e

echo "======================================"
echo "Setup Centralized Vault Configuration"
echo "======================================"
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed."
    exit 1
fi

# Get Vault configuration from user
read -p "Enter the namespace for centralized config (default: vault-system): " NAMESPACE
NAMESPACE=${NAMESPACE:-vault-system}

read -p "Enter Vault URL (e.g., https://vault.example.com:8200): " VAULT_URL
if [ -z "$VAULT_URL" ]; then
    echo "❌ Vault URL is required"
    exit 1
fi

read -sp "Enter Vault token: " VAULT_TOKEN
echo ""
if [ -z "$VAULT_TOKEN" ]; then
    echo "❌ Vault token is required"
    exit 1
fi

echo ""
echo "Configuration:"
echo "  Namespace: $NAMESPACE"
echo "  Vault URL: $VAULT_URL"
echo "  Vault Token: [hidden]"
echo ""

read -p "Continue with this configuration? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Setup cancelled."
    exit 0
fi

# Create namespace
echo "📦 Creating namespace $NAMESPACE..."
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
echo "✅ Namespace created"
echo ""

# Create TimSecretConfig
echo "📦 Creating TimSecretConfig..."
cat <<EOF | kubectl apply -f -
apiVersion: secrets.tim.operator/v1alpha1
kind: TimSecretConfig
metadata:
  name: vault-config
  namespace: $NAMESPACE
spec:
  vaultURL: "$VAULT_URL"
  vaultToken: "$VAULT_TOKEN"
EOF
echo "✅ TimSecretConfig created"
echo ""

echo "======================================"
echo "✅ Setup Complete!"
echo "======================================"
echo ""
echo "You can now create TimSecrets that reference this configuration:"
echo ""
echo "apiVersion: secrets.tim.operator/v1alpha1"
echo "kind: TimSecret"
echo "metadata:"
echo "  name: myapp-secrets"
echo "  namespace: default"
echo "spec:"
echo "  vaultConfig: vault-config"
echo "  vaultConfigNamespace: $NAMESPACE"
echo "  vaultPath: \"secret/data/myapp\""
echo "  secretName: \"myapp-secrets\""
echo ""
echo "For more examples, see examples/timsecret-with-config.yaml"


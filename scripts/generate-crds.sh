#!/bin/bash
set -e

echo "🔧 Generating CRDs from Go types..."

# Check if controller-gen is available
if ! command -v controller-gen &> /dev/null; then
    echo "📦 Installing controller-gen..."
    go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
fi

# Generate CRDs
echo "📝 Running controller-gen..."
controller-gen crd:crdVersions=v1 paths="./api/..." output:crd:artifacts:config=config/crd

# Rename generated files to our naming convention
if [ -f "config/crd/secrets.tim.operator_timsecrets.yaml" ]; then
    mv config/crd/secrets.tim.operator_timsecrets.yaml config/crd/timsecret-crd.yaml
    echo "✅ Generated timsecret-crd.yaml"
fi

if [ -f "config/crd/secrets.tim.operator_timsecretconfigs.yaml" ]; then
    mv config/crd/secrets.tim.operator_timsecretconfigs.yaml config/crd/timsecretconfig-crd.yaml
    echo "✅ Generated timsecretconfig-crd.yaml"
fi

echo "✅ CRDs generated successfully!"


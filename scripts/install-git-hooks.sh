#!/bin/bash

# Install Git Hooks for TimVault Operator

set -e

echo "======================================"
echo "Installing Git Hooks"
echo "======================================"
echo ""

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "❌ Not a git repository. Run this from the project root."
    exit 1
fi

# Configure git to use .githooks directory
echo "📝 Configuring git to use .githooks directory..."
git config core.hooksPath .githooks

# Make hooks executable
echo "🔧 Making hooks executable..."
chmod +x .githooks/*

echo ""
echo "======================================"
echo "✅ Git Hooks Installed Successfully!"
echo "======================================"
echo ""
echo "Installed hooks:"
echo "  - pre-commit: Auto-generates CRDs when API files change"
echo ""
echo "To disable hooks temporarily:"
echo "  git commit --no-verify"
echo ""
echo "To uninstall hooks:"
echo "  git config --unset core.hooksPath"


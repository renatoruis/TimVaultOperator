# Git Hooks

This directory contains Git hooks for the TimVault Operator project.

## Installation

Run the installation script from the project root:

```bash
bash scripts/install-git-hooks.sh
```

This configures Git to use hooks from this directory.

## Available Hooks

### pre-commit

Automatically regenerates CRDs from Go types when API files are changed.

**What it does:**
1. Detects if any files in `api/` were modified
2. Runs `scripts/generate-crds.sh` to regenerate CRDs
3. Automatically adds updated CRDs to your commit

**Example:**
```bash
# Edit api/v1alpha1/timsecret_types.go
vim api/v1alpha1/timsecret_types.go

# Commit - hook automatically updates CRDs
git add api/v1alpha1/timsecret_types.go
git commit -m "feat: add new field to TimSecret"

# Output:
# 🔍 Pre-commit: Checking for API changes...
# 📝 API files changed, regenerating CRDs...
# ✅ CRDs added to commit
```

## Bypassing Hooks

To skip hooks temporarily (not recommended):

```bash
git commit --no-verify
```

## Uninstalling

To remove the git hooks configuration:

```bash
git config --unset core.hooksPath
```

## Manual CRD Generation

You can also generate CRDs manually:

```bash
# Using the script
bash scripts/generate-crds.sh

# Using Make
make manifests

# Generate everything (code + CRDs)
make generate-all
```

## Troubleshooting

### Hook not running

Ensure hooks are executable:
```bash
chmod +x .githooks/*
```

Verify git configuration:
```bash
git config core.hooksPath
# Should output: .githooks
```

### controller-gen not found

The hook will automatically install controller-gen if missing, but you can install it manually:
```bash
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
```


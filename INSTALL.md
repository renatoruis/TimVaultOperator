# Installation Guide

## Quick Install

```bash
kubectl apply -f https://github.com/renatoruis/TimVaultOperator/releases/latest/download/install.yaml
```

## Installation Order

The `install.yaml` applies resources in this specific order to avoid dependency issues:

1. **CRDs** - Custom Resource Definitions must be created first
   - `timsecrets.secrets.tim.operator`
   - `timsecretconfigs.secrets.tim.operator`

2. **Namespace** - Create the operator namespace
   - `timvault-operator-system`

3. **RBAC** - Set up permissions in order:
   - ClusterRole (`timvault-operator-role`)
   - ServiceAccount (`timvault-operator`)
   - ClusterRoleBinding (`timvault-operator-rolebinding`)

4. **Deployment** - Finally deploy the operator
   - `timvault-operator-controller`

## Manual Step-by-Step Installation

If you prefer to install components individually:

```bash
# 1. Install CRDs
kubectl apply -f https://github.com/renatoruis/TimVaultOperator/releases/latest/download/timsecret-crd.yaml
kubectl apply -f https://github.com/renatoruis/TimVaultOperator/releases/latest/download/timsecretconfig-crd.yaml

# 2. Create namespace
kubectl create namespace timvault-operator-system

# 3. Install RBAC
kubectl apply -f config/rbac/role.yaml
kubectl apply -f config/rbac/service_account.yaml
kubectl apply -f config/rbac/role_binding.yaml

# 4. Deploy operator
kubectl apply -f config/manager/deployment.yaml
```

## Verification

Check that everything is running:

```bash
# Check CRDs
kubectl get crds | grep tim

# Check operator pod
kubectl get pods -n timvault-operator-system

# Check operator logs
kubectl logs -n timvault-operator-system deployment/timvault-operator-controller
```

## Troubleshooting

### Error: "namespaces not found"

This happens if resources are applied out of order. The fix is already implemented in the latest release. If you encounter this:

```bash
# Delete everything
kubectl delete -f install.yaml

# Wait a moment
sleep 5

# Reinstall
kubectl apply -f https://github.com/renatoruis/TimVaultOperator/releases/latest/download/install.yaml
```

### Pod is CrashLooping

Check the logs:
```bash
kubectl logs -n timvault-operator-system deployment/timvault-operator-controller
```

Common causes:
- Missing RBAC permissions (fixed in latest release)
- Image pull errors (check image name and registry)
- Invalid configuration

### CRDs not found

Ensure CRDs are installed:
```bash
kubectl get crd timsecrets.secrets.tim.operator
kubectl get crd timsecretconfigs.secrets.tim.operator
```

If missing, reinstall:
```bash
kubectl apply -f https://github.com/renatoruis/TimVaultOperator/releases/latest/download/timsecret-crd.yaml
kubectl apply -f https://github.com/renatoruis/TimVaultOperator/releases/latest/download/timsecretconfig-crd.yaml
```

## Uninstallation

```bash
# Quick uninstall
kubectl delete -f https://github.com/renatoruis/TimVaultOperator/releases/latest/download/install.yaml

# Or use the uninstall script
./scripts/uninstall.sh
```

## Next Steps

After installation, create your first TimSecret:

```bash
# 1. Create Vault configuration
kubectl apply -f examples/timsecretconfig-example.yaml

# 2. Create a secret sync
kubectl apply -f examples/timsecret-with-config.yaml

# 3. Verify
kubectl get timsecrets
kubectl get secrets
```

See [README.md](README.md) for more details and examples.


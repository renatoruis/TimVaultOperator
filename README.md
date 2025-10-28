# TimVault Operator

A simple and powerful Kubernetes operator exclusively designed to work with HashiCorp Vault. TimVault Operator synchronizes secrets from Vault to Kubernetes Secrets and automatically restarts deployments when secrets change.

> **Note:** This is NOT the External Secrets Operator. This is TimVault Operator - a dedicated, lightweight solution for Vault secret management.

## Features

- **Vault Integration**: Connects to HashiCorp Vault and fetches secrets from specified paths
- **Centralized Configuration**: Create `TimSecretConfig` resources to centralize Vault credentials
- **Automatic Sync**: Creates and updates Kubernetes Secrets with data from Vault
- **Deployment Restart**: Automatically restarts specified deployments when secrets are updated
- **Change Detection**: Uses hash-based change detection to avoid unnecessary restarts
- **Periodic Sync**: Re-syncs secrets every 5 minutes to keep them up to date
- **KV Support**: Works with both Vault KV v1 and KV v2 engines
- **Cross-Namespace**: Reference Vault configuration from any namespace

## Installation

### Using GitHub Release

```bash
kubectl apply -f https://github.com/renatoruis/TimVaultOperator/releases/latest/download/install.yaml
```

### Manual Installation

```bash
# Install CRDs
kubectl apply -f config/crd/timsecret-crd.yaml
kubectl apply -f config/crd/timsecretconfig-crd.yaml

# Install RBAC and Operator
kubectl apply -f config/rbac/service_account.yaml
kubectl apply -f config/rbac/role.yaml
kubectl apply -f config/rbac/role_binding.yaml
kubectl apply -f config/manager/deployment.yaml
```

### Build from Source

```bash
# Build the operator
make build

# Build Docker image
make docker-build IMG=ghcr.io/renatoruis/timvault-operator:v1.0.0

# Push to registry
make docker-push IMG=ghcr.io/renatoruis/timvault-operator:v1.0.0

# Deploy
make deploy
```

## Quick Start

### Step 1: Create Centralized Vault Configuration

Create a `TimSecretConfig` with your Vault credentials:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: vault-system
---
apiVersion: secrets.tim.operator/v1alpha1
kind: TimSecretConfig
metadata:
  name: vault-config
  namespace: vault-system
spec:
  vaultURL: "https://vault.example.com:8200"
  vaultToken: "s.xxxxxxxxxxxxxx"
```

Apply it:

```bash
kubectl apply -f examples/timsecretconfig-example.yaml
```

### Step 2: Create TimSecret

Reference the centralized config in your TimSecret:

```yaml
apiVersion: secrets.tim.operator/v1alpha1
kind: TimSecret
metadata:
  name: myapp-secrets
  namespace: default
spec:
  # Reference to TimSecretConfig
  vaultConfig: vault-config
  vaultConfigNamespace: vault-system
  
  # Path in Vault
  vaultPath: "secret/data/myapp"
  
  # Kubernetes Secret name
  secretName: "myapp-secrets"
  
  # Optional: Deployment to restart
  deploymentName: "myapp"
```

Apply it:

```bash
kubectl apply -f examples/timsecret-with-config.yaml
```

### Step 3: Use the Secret in Your Application

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
        - name: app
          image: myapp:latest
          envFrom:
            - secretRef:
                name: myapp-secrets
```

## Configuration Options

### Option 1: Using TimSecretConfig (Recommended)

Best for production environments with multiple applications:

```yaml
# Create once
apiVersion: secrets.tim.operator/v1alpha1
kind: TimSecretConfig
metadata:
  name: vault-config
  namespace: vault-system
spec:
  vaultURL: "https://vault.example.com:8200"
  vaultToken: "s.xxxxxxxxxxxxxx"
```

```yaml
# Use in multiple TimSecrets
apiVersion: secrets.tim.operator/v1alpha1
kind: TimSecret
metadata:
  name: app1-secrets
spec:
  vaultConfig: vault-config
  vaultConfigNamespace: vault-system
  vaultPath: "secret/data/app1"
  secretName: "app1-secrets"
```

**Benefits:**
- ✅ Update credentials in one place
- ✅ Better security (centralized control)
- ✅ Cleaner TimSecret definitions
- ✅ Share config across namespaces

### Option 2: Direct Configuration

Best for testing or simple deployments:

```yaml
apiVersion: secrets.tim.operator/v1alpha1
kind: TimSecret
metadata:
  name: myapp-secrets
spec:
  # Direct values
  vaultURL: "https://vault.example.com:8200"
  vaultToken: "s.xxxxxxxxxxxxxx"
  vaultPath: "secret/data/myapp"
  secretName: "myapp-secrets"
```

**Note:** Direct values override TimSecretConfig if both are specified.

## API Reference

### TimSecretConfig

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `vaultURL` | string | Yes | Vault server URL |
| `vaultToken` | string | Yes | Vault authentication token |

### TimSecret

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `vaultConfig` | string | No* | Name of TimSecretConfig to use |
| `vaultConfigNamespace` | string | No | Namespace of TimSecretConfig (defaults to TimSecret's namespace) |
| `vaultURL` | string | No* | Vault URL (direct value, overrides vaultConfig) |
| `vaultToken` | string | No* | Vault token (direct value, overrides vaultConfig) |
| `vaultPath` | string | Yes | Path in Vault where secrets are stored |
| `secretName` | string | Yes | Name of Kubernetes Secret to create |
| `deploymentName` | string | No | Deployment to restart when secrets change |
| `namespace` | string | No | Namespace for secret/deployment (defaults to TimSecret's namespace) |

\* Either `vaultConfig` or both `vaultURL` and `vaultToken` must be specified.

## Examples

All examples are available in the [`examples/`](examples/) directory:

- **`timsecretconfig-example.yaml`** - Centralized Vault configuration
- **`timsecret-with-config.yaml`** - TimSecret using centralized config
- **`timsecret-example.yaml`** - TimSecret with direct values
- **`deployment-example.yaml`** - Sample deployment using secrets

## GitHub Actions CI/CD

This project includes automated workflows:

### CI Workflow (`.github/workflows/ci.yaml`)
- Runs on every push and PR
- Go tests, linting, and formatting
- Docker image build validation
- Manifest validation

### Release Workflow (`.github/workflows/release.yaml`)
- Triggers on version tags (`v*`)
- Builds and pushes Docker image to GitHub Container Registry
- Creates GitHub release with:
  - Complete installation manifest (`install.yaml`)
  - Individual CRD files
  - Examples bundle
- Automatic release notes generation

### CRD Publishing (`.github/workflows/publish-crds.yaml`)
- Publishes CRDs on changes
- Creates artifacts for download

### Creating a Release

```bash
# Tag the release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# GitHub Actions will automatically:
# - Build and push Docker image to ghcr.io
# - Create GitHub release with manifests
# - Generate release notes
```

Users can install with:
```bash
kubectl apply -f https://github.com/renatoruis/TimVaultOperator/releases/download/v1.0.0/install.yaml
```

## Monitoring

### Check Operator Status

```bash
kubectl get pods -n timvault-operator-system
kubectl logs -n timvault-operator-system deployment/timvault-operator-controller -f
```

### Check TimSecrets

```bash
# List all TimSecrets
kubectl get timsecrets --all-namespaces

# Describe specific TimSecret
kubectl describe timsecret myapp-secrets

# Check the created Secret
kubectl get secret myapp-secrets
kubectl describe secret myapp-secrets
```

### Check TimSecretConfigs

```bash
# List configs
kubectl get timsecretconfigs -n vault-system

# Describe config
kubectl describe timsecretconfig vault-config -n vault-system
```

## Troubleshooting

### Secret Not Created

1. Check operator logs:
   ```bash
   kubectl logs -n timvault-operator-system deployment/timvault-operator-controller
   ```

2. Verify TimSecret status:
   ```bash
   kubectl describe timsecret myapp-secrets
   ```

3. Verify Vault connectivity:
   ```bash
   kubectl exec -n timvault-operator-system deployment/timvault-operator-controller -- \
     wget -O- https://vault.example.com:8200/v1/sys/health
   ```

### Config Not Found

Ensure TimSecretConfig exists in the correct namespace:

```bash
kubectl get timsecretconfig vault-config -n vault-system
```

### Deployment Not Restarting

1. Verify `deploymentName` is set in TimSecret
2. Check deployment exists: `kubectl get deployment myapp`
3. Review operator logs for errors

## Security Best Practices

1. **Store Vault Token Securely**: Use Kubernetes Secrets to store the TimSecretConfig
2. **Use RBAC**: Restrict who can create/modify TimSecretConfigs
3. **Namespace Isolation**: Keep sensitive configs in dedicated namespaces
4. **Token Rotation**: Regularly rotate Vault tokens
5. **Audit Access**: Monitor who accesses TimSecretConfigs
6. **Network Policies**: Restrict network access to Vault

## Development

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- Access to a Kubernetes cluster

### Local Development

```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Run locally (connects to current kubeconfig context)
go run cmd/main.go

# Build binary
make build

# Run linters
make fmt
make vet
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linters
5. Submit a pull request

## License

MIT License

## Support

For issues, questions, or contributions:
- Repository: https://github.com/renatoruis/TimVaultOperator
- Open an issue on GitHub
- Submit a pull request
- Check existing documentation and examples

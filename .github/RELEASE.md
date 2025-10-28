# Release Instructions

## Automated Release Process

This project uses GitHub Actions to automate releases. Here's how to create a new release:

### 1. Create a Version Tag

```bash
# Ensure you're on main branch
git checkout main
git pull origin main

# Create and push a version tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 2. GitHub Actions Will Automatically:

✅ **Build Docker Image**
- Builds multi-platform Docker image
- Pushes to GitHub Container Registry (ghcr.io)
- Tags: `v1.0.0`, `1.0`, `1`, `latest`

✅ **Create Installation Manifests**
- `install.yaml` - Complete installation in one file
- `externalsecret-crd.yaml` - ExternalSecret CRD
- `externalsecretconfig-crd.yaml` - ExternalSecretConfig CRD
- `examples.tar.gz` - All example files

✅ **Create GitHub Release**
- Automatic release notes
- Attached artifacts
- Installation instructions

### 3. Users Can Install With:

```bash
# Quick install (everything)
kubectl apply -f https://github.com/YOUR_USERNAME/ExternalSecretTimOperator/releases/download/v1.0.0/install.yaml

# Or install CRDs only
kubectl apply -f https://github.com/YOUR_USERNAME/ExternalSecretTimOperator/releases/download/v1.0.0/externalsecret-crd.yaml
kubectl apply -f https://github.com/YOUR_USERNAME/ExternalSecretTimOperator/releases/download/v1.0.0/externalsecretconfig-crd.yaml
```

## CI Workflows

### Continuous Integration (`ci.yaml`)

Runs on every push and pull request:

- ✅ Go tests with coverage
- ✅ Code formatting check (`go fmt`)
- ✅ Code linting (`go vet`)
- ✅ Docker image build validation
- ✅ Kubernetes manifest validation

### CRD Publishing (`publish-crds.yaml`)

Runs when CRDs are modified:

- ✅ Publishes CRDs as artifacts
- ✅ Optional: Publishes to GitHub Pages

## Docker Image Location

After release, the image is available at:

```
ghcr.io/YOUR_USERNAME/externalsecrettimoperator:v1.0.0
ghcr.io/YOUR_USERNAME/externalsecrettimoperator:latest
```

## Version Schema

We follow semantic versioning:

- `v1.0.0` - Major.Minor.Patch
- `v1.0.1` - Patch release (bug fixes)
- `v1.1.0` - Minor release (new features, backward compatible)
- `v2.0.0` - Major release (breaking changes)

## Pre-release Versions

For testing releases:

```bash
git tag -a v1.0.0-rc.1 -m "Release candidate 1"
git push origin v1.0.0-rc.1
```

GitHub Actions will create a pre-release.

## Manual Release (if needed)

If you need to create a release manually:

```bash
# Build Docker image
docker build -t ghcr.io/YOUR_USERNAME/externalsecrettimoperator:v1.0.0 .
docker push ghcr.io/YOUR_USERNAME/externalsecrettimoperator:v1.0.0

# Create release artifacts
mkdir -p release
cat config/crd/*.yaml config/rbac/*.yaml config/manager/*.yaml > release/install.yaml

# Create GitHub release manually
gh release create v1.0.0 \
  --title "Release v1.0.0" \
  --generate-notes \
  release/install.yaml
```

## Troubleshooting

### Release Failed

Check GitHub Actions logs:
1. Go to Actions tab in GitHub
2. Click on the failed workflow
3. Review logs for errors

Common issues:
- ❌ Docker build failed → Check Dockerfile
- ❌ Permission denied → Check GITHUB_TOKEN permissions
- ❌ Tag already exists → Delete and recreate tag

### Image Not Accessible

Make sure the GitHub Container Registry package is public:
1. Go to Package settings in GitHub
2. Change visibility to Public
3. Configure package permissions

## Repository

GitHub: https://github.com/renatoruis/TimVaultOperator

## Best Practices

1. **Test Before Release**: Always test in a dev environment first
2. **Update README**: Keep documentation in sync with releases
3. **Changelog**: Update release notes with important changes
4. **Breaking Changes**: Bump major version for breaking changes
5. **Security**: Scan images before release


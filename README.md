# Yuk - Kubernetes Image Update Controller

Yuk is a Kubernetes controller that monitors container image repositories and automatically updates Kubernetes deployments with the latest image tags, pushing changes back to Git repositories for GitOps workflows.

## Features

- Monitor AWS ECR repositories for new image tags
- Automatically update YAML configurations with latest image versions
- Push updates to GitHub repositories for GitOps tooling
- Custom Resource Definition for flexible configuration
- Decoupled from ArgoCD and FluxCD
- Comprehensive Prometheus metrics for monitoring and alerting
- Helm chart for easy deployment

## Quick Start

```bash
# Install the controller
helm install yuk ./chart/yuk

# Create a YukConfig resource
kubectl apply -f examples/yukconfig.yaml
```

## Architecture

Yuk consists of:
- Custom Resource Definition (YukConfig) for configuration
- Controller that watches for changes in configured repositories
- GitHub integration for pushing updates
- AWS ECR integration for monitoring image repositories

## Configuration

See [examples/](examples/) for sample configurations.

## Development

```bash
# Run tests
make test

# Build
make build

# Run locally
make run
```

## License

MIT License
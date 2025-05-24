# Getting Started with Yuk

This guide will help you get started with Yuk, a Kubernetes controller for automated image updates in GitOps workflows.

## Prerequisites

- Kubernetes cluster (1.20+)
- Helm 3.0+
- kubectl configured for your cluster
- AWS ECR repository with container images
- GitHub repository for storing Kubernetes manifests

## Installation

### 1. Install Yuk using Helm

```bash
# Add the Helm repository (when available)
helm repo add yuk https://charts.rebelops.io

# Install Yuk
helm install yuk yuk/yuk --namespace yuk-system --create-namespace
```

### 2. Install from source

```bash
# Clone the repository
git clone https://github.com/rebelopsio/yuk.git
cd yuk

# Install CRDs
make install

# Deploy using Helm
helm install yuk ./chart/yuk --namespace yuk-system --create-namespace
```

## Configuration

### 1. Create a GitHub Personal Access Token

1. Go to GitHub Settings > Developer settings > Personal access tokens
2. Generate a new token with `repo` scope
3. Save the token securely

### 2. Create a Kubernetes Secret

```bash
kubectl create secret generic github-token \
  --from-literal=token=your-github-token \
  --namespace default
```

### 3. Configure AWS IAM (for ECR access)

If using IRSA (IAM Roles for Service Accounts):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchCheckLayerAvailability",
        "ecr:GetDownloadUrlForLayer",
        "ecr:BatchGetImage",
        "ecr:DescribeRepositories",
        "ecr:DescribeImages"
      ],
      "Resource": "*"
    }
  ]
}
```

### 4. Create a YukConfig Resource

```yaml
apiVersion: yuk.rebelops.io/v1
kind: YukConfig
metadata:
  name: my-app-config
  namespace: default
spec:
  repository:
    type: ecr
    ecr:
      region: us-east-1
      repositoryName: my-app
      auth:
        useIRSA: true
  
  git:
    repository: https://github.com/myorg/k8s-manifests.git
    branch: main
    email: yuk@myorg.com
    name: Yuk Controller
    auth:
      personalAccessTokenRef:
        name: github-token
        key: token
  
  updateTargets:
    - file: apps/my-app/deployment.yaml
      yamlPath: spec.template.spec.containers[0].image
      imageTagOnly: true
```

Apply the configuration:

```bash
kubectl apply -f yukconfig.yaml
```

## Verification

### 1. Check YukConfig Status

```bash
kubectl get yuk
kubectl describe yuk my-app-config
```

### 2. View Logs

```bash
kubectl logs -n yuk-system deployment/yuk
```

### 3. Monitor Updates

Yuk will automatically:
1. Check ECR for new image tags every 5 minutes (configurable)
2. Update the specified YAML files when new tags are found
3. Commit and push changes to your Git repository
4. Update the YukConfig status with the latest information

## Advanced Configuration

### Tag Filtering

Use regex patterns to filter which tags to consider:

```yaml
spec:
  repository:
    ecr:
      tagFilter: "^v[0-9]+\\.[0-9]+\\.[0-9]+$"  # Only semantic versions
```

### Multiple Update Targets

Update multiple files or keys:

```yaml
spec:
  updateTargets:
    - file: apps/my-app/deployment.yaml
      yamlPath: spec.template.spec.containers[0].image
      imageTagOnly: true
    - file: apps/my-app/kustomization.yaml
      yamlPath: images[0].newTag
      imageTagOnly: false
```

### Custom Check Intervals

```yaml
spec:
  checkInterval: 10m  # Check every 10 minutes
```

## Troubleshooting

### Common Issues

1. **Authentication Errors**: Ensure your GitHub token has proper permissions and the secret is correctly created.

2. **AWS ECR Access**: Verify IAM permissions and IRSA configuration.

3. **YAML Path Errors**: Check that your YAML paths are correct using the examples.

4. **Git Push Failures**: Ensure the repository exists and the token has write access.

### Getting Help

- Check the controller logs: `kubectl logs -n yuk-system deployment/yuk`
- View YukConfig status: `kubectl describe yuk <name>`
- Open an issue: https://github.com/rebelopsio/yuk/issues
# API Reference

This document describes the Yuk API types and their configuration options.

## YukConfig

YukConfig is the main configuration resource for Yuk. It defines which repository to monitor, how to access Git repositories, and what files to update.

### YukConfigSpec

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `repository` | [RepositoryConfig](#repositoryconfig) | Configuration for the repository to monitor | Yes |
| `git` | [GitConfig](#gitconfig) | Configuration for Git operations | Yes |
| `updateTargets` | [][UpdateTarget](#updatetarget) | List of files and keys to update | Yes |
| `checkInterval` | `metav1.Duration` | How often to check for updates (default: 5m) | No |
| `disabled` | `bool` | Whether this configuration is disabled | No |

### RepositoryConfig

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `type` | `string` | Type of repository ("ecr") | Yes |
| `ecr` | [ECRConfig](#ecrconfig) | ECR-specific configuration | When type is "ecr" |

### ECRConfig

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `region` | `string` | AWS region where the ECR repository is located | Yes |
| `repositoryName` | `string` | Name of the ECR repository | Yes |
| `tagFilter` | `string` | Regex pattern to filter tags | No |
| `auth` | [ECRAuthConfig](#ecrauthconfig) | Authentication configuration | No |

### ECRAuthConfig

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `useIRSA` | `bool` | Whether to use IAM Roles for Service Accounts | No |
| `accessKeyID` | `string` | AWS Access Key ID (if not using IRSA) | No |
| `secretAccessKeyRef` | [SecretKeySelector](#secretkeyselector) | Reference to secret containing AWS Secret Access Key | No |

### GitConfig

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `repository` | `string` | Git repository URL | Yes |
| `branch` | `string` | Branch to update (default: "main") | No |
| `auth` | [GitAuthConfig](#gitauthconfig) | Authentication configuration | Yes |
| `commitMessage` | `string` | Commit message template | No |
| `email` | `string` | Email for git commits | Yes |
| `name` | `string` | Name for git commits | Yes |

### GitAuthConfig

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `personalAccessTokenRef` | [SecretKeySelector](#secretkeyselector) | Reference to GitHub Personal Access Token | No |
| `sshKeyRef` | [SecretKeySelector](#secretkeyselector) | Reference to SSH private key | No |

### UpdateTarget

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `file` | `string` | Path to file in Git repository | Yes |
| `yamlPath` | `string` | YAML key path to update | Yes |
| `imageTagOnly` | `bool` | Whether to update only the tag part of an image reference | No |

### SecretKeySelector

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `name` | `string` | Name of the secret | Yes |
| `key` | `string` | Key within the secret | Yes |

### YukConfigStatus

| Field | Type | Description |
|-------|------|-------------|
| `lastChecked` | `metav1.Time` | Timestamp of last repository check |
| `lastUpdate` | `metav1.Time` | Timestamp of last successful update |
| `currentTag` | `string` | Current tag being monitored |
| `latestTag` | `string` | Latest tag found in repository |
| `conditions` | `[]metav1.Condition` | Current state conditions |
| `observedGeneration` | `int64` | Observed generation of the resource |

## YAML Path Format

The `yamlPath` field uses a dot-notation format to specify keys in YAML files:

### Examples

- `spec.template.spec.containers[0].image` - Navigate to the first container's image
- `metadata.labels.version` - Update a label value
- `data.config` - Update a ConfigMap data field
- `spec.replicas` - Update replica count

### Array Indexing

Use square brackets with zero-based indexing:
- `containers[0]` - First container
- `volumes[1]` - Second volume
- `env[2]` - Third environment variable

### Image Tag Only Updates

When `imageTagOnly: true`, Yuk will:
1. Parse the current image reference (e.g., `registry/image:tag`)
2. Replace only the tag portion with the new tag
3. Preserve the registry and image name

Example:
- Current: `docker.io/nginx:1.20`
- New tag: `1.21`
- Result: `docker.io/nginx:1.21`

## Conditions

YukConfig resources use standard Kubernetes conditions to report status:

### Condition Types

- `Ready` - Whether the configuration is ready and functioning
- `RepositoryAccessible` - Whether the repository can be accessed
- `GitAccessible` - Whether the Git repository can be accessed

### Condition Reasons

- `Synchronized` - Successfully synchronized with repository
- `RepositoryError` - Error accessing the repository
- `GitError` - Error with Git operations
- `UpdateError` - Error updating files
- `AuthenticationError` - Authentication failure
/*
MIT License

Copyright (c) 2024 Yuk Contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// YukConfigSpec defines the desired state of YukConfig
type YukConfigSpec struct {
	// Repository defines the configuration for the repository to monitor
	Repository RepositoryConfig `json:"repository"`

	// Git defines the configuration for Git operations
	Git GitConfig `json:"git"`

	// UpdateTargets defines what files and keys to update
	UpdateTargets []UpdateTarget `json:"updateTargets"`

	// CheckInterval defines how often to check for updates (default: 5m)
	CheckInterval *metav1.Duration `json:"checkInterval,omitempty"`

	// Disabled can be used to temporarily disable this configuration
	Disabled bool `json:"disabled,omitempty"`
}

// RepositoryConfig defines the repository to monitor
type RepositoryConfig struct {
	// Type defines the type of repository (currently only "ecr")
	Type string `json:"type"`

	// ECR configuration (when type is "ecr")
	ECR *ECRConfig `json:"ecr,omitempty"`
}

// ECRConfig defines AWS ECR specific configuration
type ECRConfig struct {
	// Region is the AWS region where the ECR repository is located
	Region string `json:"region"`

	// RepositoryName is the name of the ECR repository
	RepositoryName string `json:"repositoryName"`

	// TagFilter allows filtering tags (regex pattern)
	TagFilter string `json:"tagFilter,omitempty"`

	// Authentication configuration
	Auth ECRAuthConfig `json:"auth,omitempty"`
}

// ECRAuthConfig defines authentication for ECR
type ECRAuthConfig struct {
	// UseIRSA indicates whether to use IAM Roles for Service Accounts
	UseIRSA bool `json:"useIRSA,omitempty"`

	// AccessKeyID for ECR authentication (if not using IRSA)
	AccessKeyID string `json:"accessKeyID,omitempty"`

	// SecretAccessKey for ECR authentication (stored in a secret)
	SecretAccessKeyRef *SecretKeySelector `json:"secretAccessKeyRef,omitempty"`
}

// GitConfig defines Git repository configuration
type GitConfig struct {
	// Repository URL (e.g., https://github.com/owner/repo.git)
	Repository string `json:"repository"`

	// Branch to update (default: main)
	Branch string `json:"branch,omitempty"`

	// Authentication configuration
	Auth GitAuthConfig `json:"auth"`

	// CommitMessage template for updates
	CommitMessage string `json:"commitMessage,omitempty"`

	// Email for git commits
	Email string `json:"email"`

	// Name for git commits
	Name string `json:"name"`
}

// GitAuthConfig defines authentication for Git operations
type GitAuthConfig struct {
	// PersonalAccessToken reference for GitHub authentication
	PersonalAccessTokenRef *SecretKeySelector `json:"personalAccessTokenRef,omitempty"`

	// SSHKey reference for SSH authentication
	SSHKeyRef *SecretKeySelector `json:"sshKeyRef,omitempty"`
}

// UpdateTarget defines what to update in the Git repository
type UpdateTarget struct {
	// File path in the Git repository
	File string `json:"file"`

	// YAMLPath defines the YAML key to update (e.g., "spec.template.spec.containers[0].image")
	YAMLPath string `json:"yamlPath"`

	// ImageTagOnly indicates whether to update only the tag part of an image reference
	ImageTagOnly bool `json:"imageTagOnly,omitempty"`
}

// SecretKeySelector selects a key of a Secret
type SecretKeySelector struct {
	// The name of the secret in the pod's namespace to select from
	Name string `json:"name"`

	// The key of the secret to select from
	Key string `json:"key"`
}

// YukConfigStatus defines the observed state of YukConfig
type YukConfigStatus struct {
	// LastChecked is the timestamp of the last repository check
	LastChecked *metav1.Time `json:"lastChecked,omitempty"`

	// LastUpdate is the timestamp of the last successful update
	LastUpdate *metav1.Time `json:"lastUpdate,omitempty"`

	// CurrentTag is the current tag/version being monitored
	CurrentTag string `json:"currentTag,omitempty"`

	// LatestTag is the latest tag found in the repository
	LatestTag string `json:"latestTag,omitempty"`

	// Conditions represent the latest available observations of the YukConfig's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed YukConfig
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Namespaced,shortName=yuk
//+kubebuilder:printcolumn:name="Repository",type="string",JSONPath=".spec.repository.ecr.repositoryName"
//+kubebuilder:printcolumn:name="Current Tag",type="string",JSONPath=".status.currentTag"
//+kubebuilder:printcolumn:name="Latest Tag",type="string",JSONPath=".status.latestTag"
//+kubebuilder:printcolumn:name="Last Update",type="date",JSONPath=".status.lastUpdate"

// YukConfig is the Schema for the yukconfigs API
type YukConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   YukConfigSpec   `json:"spec,omitempty"`
	Status YukConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// YukConfigList contains a list of YukConfig
type YukConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []YukConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&YukConfig{}, &YukConfigList{})
}

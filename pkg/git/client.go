/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package git

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	yukv1 "github.com/rebelopsio/yuk/apis/yuk/v1"
)

// Client provides operations for interacting with Git repositories
type Client struct {
	config yukv1.GitConfig
}

// NewClient creates a new Git client with the specified configuration
func NewClient(config yukv1.GitConfig) *Client {
	return &Client{
		config: config,
	}
}

// Clone clones the repository to a temporary directory
func (c *Client) Clone(ctx context.Context) (string, error) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "yuk-git-")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	// Determine the repository URL with authentication
	repoURL, err := c.getAuthenticatedRepoURL()
	if err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to get authenticated repository URL: %w", err)
	}

	// Clone the repository
	branch := c.config.Branch
	if branch == "" {
		branch = "main"
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--single-branch", "--branch", branch, repoURL, tmpDir)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to clone repository: %w, output: %s", err, output)
	}

	// Configure git user for commits
	if err := c.configureGitUser(tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return "", fmt.Errorf("failed to configure git user: %w", err)
	}

	return tmpDir, nil
}

// CommitAndPush commits changes and pushes them to the remote repository
func (c *Client) CommitAndPush(ctx context.Context, repoPath, commitMessage string) error {
	// Add all changes
	cmd := exec.CommandContext(ctx, "git", "add", ".")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add changes: %w, output: %s", err, output)
	}

	// Check if there are changes to commit
	cmd = exec.CommandContext(ctx, "git", "diff", "--cached", "--quiet")
	cmd.Dir = repoPath
	if err := cmd.Run(); err == nil {
		// No changes to commit
		return nil
	}

	// Commit changes
	cmd = exec.CommandContext(ctx, "git", "commit", "-m", commitMessage)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to commit changes: %w, output: %s", err, output)
	}

	// Push changes
	branch := c.config.Branch
	if branch == "" {
		branch = "main"
	}

	cmd = exec.CommandContext(ctx, "git", "push", "origin", branch)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push changes: %w, output: %s", err, output)
	}

	return nil
}

// Cleanup removes the temporary repository directory
func (c *Client) Cleanup(repoPath string) {
	os.RemoveAll(repoPath)
}

// getAuthenticatedRepoURL returns the repository URL with authentication credentials
func (c *Client) getAuthenticatedRepoURL() (string, error) {
	repoURL := c.config.Repository

	// If using personal access token for GitHub
	if c.config.Auth.PersonalAccessTokenRef != nil {
		// In a real implementation, you would retrieve the token from the Kubernetes secret
		// For now, we'll assume the token is provided via environment variable
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return "", fmt.Errorf("GitHub token not found in environment")
		}

		// Convert https://github.com/owner/repo.git to https://token@github.com/owner/repo.git
		if strings.HasPrefix(repoURL, "https://github.com/") {
			repoURL = strings.Replace(repoURL, "https://github.com/", fmt.Sprintf("https://%s@github.com/", token), 1)
		}
	}

	return repoURL, nil
}

// configureGitUser configures the git user name and email for commits
func (c *Client) configureGitUser(repoPath string) error {
	// Set user name
	cmd := exec.Command("git", "config", "user.name", c.config.Name)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set git user name: %w, output: %s", err, output)
	}

	// Set user email
	cmd = exec.Command("git", "config", "user.email", c.config.Email)
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set git user email: %w, output: %s", err, output)
	}

	return nil
}

// GetLastCommitHash returns the hash of the last commit
func (c *Client) GetLastCommitHash(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get last commit hash: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetFileContent reads the content of a file in the repository
func (c *Client) GetFileContent(repoPath, filePath string) ([]byte, error) {
	fullPath := filepath.Join(repoPath, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return content, nil
}

// WriteFileContent writes content to a file in the repository
func (c *Client) WriteFileContent(repoPath, filePath string, content []byte) error {
	fullPath := filepath.Join(repoPath, filePath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory for file %s: %w", filePath, err)
	}

	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

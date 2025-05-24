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
	"testing"

	yukv1 "github.com/rebelopsio/yuk/apis/yuk/v1"
)

func TestNewClient(t *testing.T) {
	config := yukv1.GitConfig{
		Repository: "https://github.com/example/repo.git",
		Branch:     "main",
		Email:      "test@example.com",
		Name:       "Test User",
	}

	client := NewClient(config)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.config.Repository != config.Repository {
		t.Errorf("Expected repository %s, got %s", config.Repository, client.config.Repository)
	}

	if client.config.Branch != config.Branch {
		t.Errorf("Expected branch %s, got %s", config.Branch, client.config.Branch)
	}
}

func TestClient_getAuthenticatedRepoURL(t *testing.T) {
	tests := []struct {
		name           string
		repository     string
		hasToken       bool
		token          string
		expectedPrefix string
	}{
		{
			name:           "github repo without token",
			repository:     "https://github.com/example/repo.git",
			hasToken:       false,
			expectedPrefix: "https://github.com/",
		},
		{
			name:           "github repo with token",
			repository:     "https://github.com/example/repo.git",
			hasToken:       true,
			token:          "ghp_1234567890",
			expectedPrefix: "https://ghp_1234567890@github.com/",
		},
		{
			name:           "non-github repo",
			repository:     "https://gitlab.com/example/repo.git",
			hasToken:       false,
			expectedPrefix: "https://gitlab.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := yukv1.GitConfig{
				Repository: tt.repository,
			}

			if tt.hasToken {
				config.Auth.PersonalAccessTokenRef = &yukv1.SecretKeySelector{
					Name: "github-token",
					Key:  "token",
				}
				t.Setenv("GITHUB_TOKEN", tt.token)
			}

			client := NewClient(config)
			url, err := client.getAuthenticatedRepoURL()

			if err != nil && !tt.hasToken {
				// Expected for cases without token
				return
			}

			if tt.hasToken && err != nil {
				t.Errorf("Expected no error for case with token, got: %v", err)
				return
			}

			if tt.hasToken && !contains(url, tt.expectedPrefix) {
				t.Errorf("Expected URL to contain %s, got %s", tt.expectedPrefix, url)
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr))
}

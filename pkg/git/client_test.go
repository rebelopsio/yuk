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

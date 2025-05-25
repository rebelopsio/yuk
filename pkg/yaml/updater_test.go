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

package yaml

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewUpdater(t *testing.T) {
	updater := NewUpdater()
	if updater == nil {
		t.Fatal("Expected updater to be created, got nil")
	}
}

func TestUpdater_ParsePath(t *testing.T) {
	updater := NewUpdater()

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple path",
			path:     "spec.template.spec",
			expected: []string{"spec", "template", "spec"},
		},
		{
			name:     "path with array index",
			path:     "spec.containers[0].image",
			expected: []string{"spec", "containers", "0", "image"},
		},
		{
			name:     "complex path",
			path:     "spec.template.spec.containers[0].image",
			expected: []string{"spec", "template", "spec", "containers", "0", "image"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updater.parsePath(tt.path)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parts, got %d", len(tt.expected), len(result))
				return
			}
			for i, part := range result {
				if part != tt.expected[i] {
					t.Errorf("Expected part %d to be %s, got %s", i, tt.expected[i], part)
				}
			}
		})
	}
}

func TestUpdater_UpdateImageTag(t *testing.T) {
	updater := NewUpdater()

	tests := []struct {
		name         string
		currentImage string
		newTag       string
		expected     string
	}{
		{
			name:         "simple image with tag",
			currentImage: "nginx:1.20",
			newTag:       "1.21",
			expected:     "nginx:1.21",
		},
		{
			name:         "registry with image and tag",
			currentImage: "docker.io/nginx:1.20",
			newTag:       "1.21",
			expected:     "docker.io/nginx:1.21",
		},
		{
			name:         "ecr image with tag",
			currentImage: "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:v1.0.0",
			newTag:       "v1.1.0",
			expected:     "123456789012.dkr.ecr.us-east-1.amazonaws.com/my-app:v1.1.0",
		},
		{
			name:         "image without tag",
			currentImage: "nginx",
			newTag:       "1.21",
			expected:     "nginx:1.21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updater.updateImageTag(tt.currentImage, tt.newTag)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestUpdater_ValidateYAMLPath(t *testing.T) {
	updater := NewUpdater()

	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{
			name:      "valid simple path",
			path:      "spec.template.spec",
			shouldErr: false,
		},
		{
			name:      "valid path with array",
			path:      "spec.containers[0].image",
			shouldErr: false,
		},
		{
			name:      "empty path",
			path:      "",
			shouldErr: true,
		},
		{
			name:      "invalid characters",
			path:      "spec.template-spec",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := updater.ValidateYAMLPath(tt.path)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestUpdater_UpdateYAMLPath(t *testing.T) {
	updater := NewUpdater()

	// Create a temporary YAML file for testing
	yamlContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: app
        image: nginx:1.20
        ports:
        - containerPort: 80
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test updating the image tag
	err := updater.UpdateYAMLPath(tmpFile, "spec.template.spec.containers[0].image", "nginx:1.21", false)
	if err != nil {
		t.Fatalf("Failed to update YAML path: %v", err)
	}

	// Verify the update
	value, err := updater.GetValueAtPath(tmpFile, "spec.template.spec.containers[0].image")
	if err != nil {
		t.Fatalf("Failed to get value at path: %v", err)
	}

	if value != "nginx:1.21" {
		t.Errorf("Expected nginx:1.21, got %v", value)
	}
}

func TestUpdater_UpdateYAMLPath_ImageTagOnly(t *testing.T) {
	updater := NewUpdater()

	// Create a temporary YAML file for testing
	yamlContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: app
        image: docker.io/nginx:1.20
        ports:
        - containerPort: 80
`

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.yaml")
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test updating only the image tag
	err := updater.UpdateYAMLPath(tmpFile, "spec.template.spec.containers[0].image", "1.21", true)
	if err != nil {
		t.Fatalf("Failed to update YAML path: %v", err)
	}

	// Verify the update preserved the registry
	value, err := updater.GetValueAtPath(tmpFile, "spec.template.spec.containers[0].image")
	if err != nil {
		t.Fatalf("Failed to get value at path: %v", err)
	}

	if value != "docker.io/nginx:1.21" {
		t.Errorf("Expected docker.io/nginx:1.21, got %v", value)
	}
}

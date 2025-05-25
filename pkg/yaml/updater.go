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
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Updater provides functionality to update YAML files
type Updater struct{}

// NewUpdater creates a new YAML updater
func NewUpdater() *Updater {
	return &Updater{}
}

// UpdateYAMLPath updates a specific path in a YAML file with a new value
func (u *Updater) UpdateYAMLPath(filePath, yamlPath, newValue string, imageTagOnly bool) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse YAML
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return fmt.Errorf("failed to parse YAML in file %s: %w", filePath, err)
	}

	// Update the value at the specified path
	if err := u.updateValueAtPath(yamlData, yamlPath, newValue, imageTagOnly); err != nil {
		return fmt.Errorf("failed to update YAML path %s in file %s: %w", yamlPath, filePath, err)
	}

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("failed to marshal updated YAML for file %s: %w", filePath, err)
	}

	// Write back to file
	if err := os.WriteFile(filePath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated YAML to file %s: %w", filePath, err)
	}

	return nil
}

// updateValueAtPath updates a value at a specific path in the YAML structure
func (u *Updater) updateValueAtPath(data interface{}, path, newValue string, imageTagOnly bool) error {
	pathParts := u.parsePath(path)

	current := data
	for i, part := range pathParts {
		if i == len(pathParts)-1 {
			// Last part - update the value
			return u.setValue(current, part, newValue, imageTagOnly)
		}

		// Navigate to the next level
		next, err := u.getValue(current, part)
		if err != nil {
			return fmt.Errorf("failed to navigate to path part '%s': %w", part, err)
		}
		current = next
	}

	return nil
}

// parsePath parses a YAML path like "spec.template.spec.containers[0].image" into parts
func (u *Updater) parsePath(path string) []string {
	// Handle array indices like containers[0]
	arrayRegex := regexp.MustCompile(`(\w+)\[(\d+)\]`)
	path = arrayRegex.ReplaceAllString(path, "$1.$2")

	return strings.Split(path, ".")
}

// getValue gets a value from a YAML structure at a specific key/index
func (u *Updater) getValue(data interface{}, key string) (interface{}, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		if value, exists := v[key]; exists {
			return value, nil
		}
		return nil, fmt.Errorf("key '%s' not found in map", key)

	case []interface{}:
		index, err := strconv.Atoi(key)
		if err != nil {
			return nil, fmt.Errorf("invalid array index '%s': %w", key, err)
		}
		if index < 0 || index >= len(v) {
			return nil, fmt.Errorf("array index %d out of bounds (length: %d)", index, len(v))
		}
		return v[index], nil

	default:
		return nil, fmt.Errorf("cannot navigate into non-map/non-array type: %T", data)
	}
}

// setValue sets a value in a YAML structure at a specific key/index
func (u *Updater) setValue(data interface{}, key, newValue string, imageTagOnly bool) error {
	switch v := data.(type) {
	case map[string]interface{}:
		if imageTagOnly {
			// If updating only the tag part of an image reference
			currentValue, exists := v[key]
			if exists {
				if currentStr, ok := currentValue.(string); ok {
					updatedValue := u.updateImageTag(currentStr, newValue)
					v[key] = updatedValue
					return nil
				}
			}
		}
		v[key] = newValue
		return nil

	case []interface{}:
		index, err := strconv.Atoi(key)
		if err != nil {
			return fmt.Errorf("invalid array index '%s': %w", key, err)
		}
		if index < 0 || index >= len(v) {
			return fmt.Errorf("array index %d out of bounds (length: %d)", index, len(v))
		}

		if imageTagOnly {
			// If updating only the tag part of an image reference
			if currentStr, ok := v[index].(string); ok {
				updatedValue := u.updateImageTag(currentStr, newValue)
				v[index] = updatedValue
				return nil
			}
		}
		v[index] = newValue
		return nil

	default:
		return fmt.Errorf("cannot set value in non-map/non-array type: %T", data)
	}
}

// updateImageTag updates only the tag portion of a container image reference
func (u *Updater) updateImageTag(currentImage, newTag string) string {
	// Handle formats like:
	// - image:tag -> image:newTag
	// - registry/image:tag -> registry/image:newTag
	// - registry/namespace/image:tag -> registry/namespace/image:newTag

	parts := strings.Split(currentImage, ":")
	if len(parts) >= 2 {
		// Replace the last part (tag) with the new tag
		parts[len(parts)-1] = newTag
		return strings.Join(parts, ":")
	}

	// If no tag exists, append it
	return currentImage + ":" + newTag
}

// ValidateYAMLPath validates that a YAML path is correctly formatted
func (u *Updater) ValidateYAMLPath(path string) error {
	if path == "" {
		return fmt.Errorf("YAML path cannot be empty")
	}

	// Basic validation - check for valid path format
	pathRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*|\[\d+\])*$`)
	if !pathRegex.MatchString(path) {
		return fmt.Errorf("invalid YAML path format: %s", path)
	}

	return nil
}

// GetValueAtPath retrieves a value at a specific YAML path (useful for validation)
func (u *Updater) GetValueAtPath(filePath, yamlPath string) (interface{}, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Parse YAML
	var yamlData interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in file %s: %w", filePath, err)
	}

	// Navigate to the specified path
	pathParts := u.parsePath(yamlPath)
	current := yamlData

	for _, part := range pathParts {
		next, err := u.getValue(current, part)
		if err != nil {
			return nil, fmt.Errorf("failed to navigate to path part '%s': %w", part, err)
		}
		current = next
	}

	return current, nil
}

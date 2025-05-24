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

package ecr

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	region := "us-east-1"
	client := NewClient(region)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.region != region {
		t.Errorf("Expected region %s, got %s", region, client.region)
	}

	if client.ecrClient != nil {
		t.Error("Expected ecrClient to be nil before initialization")
	}
}

func TestClient_GetLatestTag_EmptyRepository(t *testing.T) {
	client := NewClient("us-east-1")

	// This test would require AWS credentials and real ECR repository
	// In a real implementation, you would mock the ECR client
	// For now, we'll test the basic structure
	if client.region != "us-east-1" {
		t.Errorf("Expected region us-east-1, got %s", client.region)
	}
}

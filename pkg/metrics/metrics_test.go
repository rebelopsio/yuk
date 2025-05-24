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

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRegisterMetrics(t *testing.T) {
	// Test that RegisterMetrics doesn't panic
	RegisterMetrics()

	// Since metrics are registered globally, we can't easily test the registration
	// without affecting other tests. Instead, we'll just verify the function runs
	// without panicking, which is the main concern.

	// We can test that the metric objects exist and are properly configured
	if ReconciliationDuration == nil {
		t.Error("ReconciliationDuration metric is nil")
	}

	if ReconciliationTotal == nil {
		t.Error("ReconciliationTotal metric is nil")
	}

	if RepositoryChecks == nil {
		t.Error("RepositoryChecks metric is nil")
	}

	if GitOperations == nil {
		t.Error("GitOperations metric is nil")
	}
}

func TestMetricConstants(t *testing.T) {
	// Test ReconciliationResult constants
	if ReconciliationSuccess != "success" {
		t.Errorf("Expected ReconciliationSuccess to be 'success', got %s", ReconciliationSuccess)
	}

	if ReconciliationError != "error" {
		t.Errorf("Expected ReconciliationError to be 'error', got %s", ReconciliationError)
	}

	if ReconciliationSkipped != "skipped" {
		t.Errorf("Expected ReconciliationSkipped to be 'skipped', got %s", ReconciliationSkipped)
	}

	// Test GitOperationType constants
	if GitOperationClone != "clone" {
		t.Errorf("Expected GitOperationClone to be 'clone', got %s", GitOperationClone)
	}

	if GitOperationCommit != "commit" {
		t.Errorf("Expected GitOperationCommit to be 'commit', got %s", GitOperationCommit)
	}

	if GitOperationPush != "push" {
		t.Errorf("Expected GitOperationPush to be 'push', got %s", GitOperationPush)
	}

	// Test ErrorType constants
	if ErrorTypeRepository != "repository" {
		t.Errorf("Expected ErrorTypeRepository to be 'repository', got %s", ErrorTypeRepository)
	}

	if ErrorTypeGit != "git" {
		t.Errorf("Expected ErrorTypeGit to be 'git', got %s", ErrorTypeGit)
	}

	if ErrorTypeYAML != "yaml" {
		t.Errorf("Expected ErrorTypeYAML to be 'yaml', got %s", ErrorTypeYAML)
	}
}

func TestMetricLabels(t *testing.T) {
	// Test that we can create metrics with expected labels
	ReconciliationTotal.With(prometheus.Labels{
		"namespace": "test-ns",
		"name":      "test-config",
		"result":    string(ReconciliationSuccess),
	}).Inc()

	// Verify the metric was recorded
	value := testutil.ToFloat64(ReconciliationTotal.With(prometheus.Labels{
		"namespace": "test-ns",
		"name":      "test-config",
		"result":    string(ReconciliationSuccess),
	}))

	if value != 1.0 {
		t.Errorf("Expected metric value to be 1.0, got %f", value)
	}
}

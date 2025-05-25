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

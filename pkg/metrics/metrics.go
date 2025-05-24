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
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// ReconciliationDuration tracks the time taken for reconciliation
	ReconciliationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yuk_controller_reconciliation_duration_seconds",
			Help:    "Time taken for YukConfig reconciliation",
			Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"namespace", "name", "result"},
	)

	// ReconciliationTotal tracks the total number of reconciliations
	ReconciliationTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yuk_controller_reconciliation_total",
			Help: "Total number of reconciliations performed",
		},
		[]string{"namespace", "name", "result"},
	)

	// RepositoryChecks tracks repository check operations
	RepositoryChecks = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yuk_repository_checks_total",
			Help: "Total number of repository checks performed",
		},
		[]string{"repository_type", "repository_name", "result"},
	)

	// RepositoryCheckDuration tracks time taken for repository checks
	RepositoryCheckDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yuk_repository_check_duration_seconds",
			Help:    "Time taken for repository checks",
			Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0},
		},
		[]string{"repository_type", "repository_name"},
	)

	// GitOperations tracks Git operations
	GitOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yuk_git_operations_total",
			Help: "Total number of Git operations performed",
		},
		[]string{"operation", "repository", "result"},
	)

	// GitOperationDuration tracks time taken for Git operations
	GitOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "yuk_git_operation_duration_seconds",
			Help:    "Time taken for Git operations",
			Buckets: []float64{0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0},
		},
		[]string{"operation", "repository"},
	)

	// UpdatesPerformed tracks successful updates
	UpdatesPerformed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yuk_updates_performed_total",
			Help: "Total number of successful updates performed",
		},
		[]string{"namespace", "name", "repository_type", "repository_name"},
	)

	// FilesUpdated tracks the number of files updated
	FilesUpdated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yuk_files_updated_total",
			Help: "Total number of files updated",
		},
		[]string{"namespace", "name", "file_path"},
	)

	// CurrentVersion tracks the current version being monitored
	CurrentVersion = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yuk_current_version_info",
			Help: "Information about the current version being monitored (value is always 1)",
		},
		[]string{"namespace", "name", "repository_name", "current_tag", "latest_tag"},
	)

	// ConfigStatus tracks the status of YukConfig resources
	ConfigStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yuk_config_status",
			Help: "Status of YukConfig resources (1=ready, 0=not ready)",
		},
		[]string{"namespace", "name", "condition_type"},
	)

	// LastCheckTimestamp tracks when repositories were last checked
	LastCheckTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yuk_last_check_timestamp_seconds",
			Help: "Timestamp of the last repository check",
		},
		[]string{"namespace", "name", "repository_name"},
	)

	// LastUpdateTimestamp tracks when updates were last performed
	LastUpdateTimestamp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yuk_last_update_timestamp_seconds",
			Help: "Timestamp of the last successful update",
		},
		[]string{"namespace", "name", "repository_name"},
	)

	// QueueDepth tracks the controller's work queue depth
	QueueDepth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "yuk_controller_queue_depth",
			Help: "Current depth of the controller work queue",
		},
		[]string{"controller"},
	)

	// ErrorsTotal tracks various types of errors
	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "yuk_errors_total",
			Help: "Total number of errors encountered",
		},
		[]string{"error_type", "namespace", "name"},
	)
)

// RegisterMetrics registers all Yuk metrics with the controller-runtime metrics registry
func RegisterMetrics() {
	metrics.Registry.MustRegister(
		ReconciliationDuration,
		ReconciliationTotal,
		RepositoryChecks,
		RepositoryCheckDuration,
		GitOperations,
		GitOperationDuration,
		UpdatesPerformed,
		FilesUpdated,
		CurrentVersion,
		ConfigStatus,
		LastCheckTimestamp,
		LastUpdateTimestamp,
		QueueDepth,
		ErrorsTotal,
	)
}

// ReconciliationResult represents the result of a reconciliation
type ReconciliationResult string

const (
	ReconciliationSuccess ReconciliationResult = "success"
	ReconciliationError   ReconciliationResult = "error"
	ReconciliationSkipped ReconciliationResult = "skipped"
)

// RepositoryCheckResult represents the result of a repository check
type RepositoryCheckResult string

const (
	RepositoryCheckSuccess RepositoryCheckResult = "success"
	RepositoryCheckError   RepositoryCheckResult = "error"
)

// GitOperationType represents different types of Git operations
type GitOperationType string

const (
	GitOperationClone  GitOperationType = "clone"
	GitOperationCommit GitOperationType = "commit"
	GitOperationPush   GitOperationType = "push"
)

// GitOperationResult represents the result of a Git operation
type GitOperationResult string

const (
	GitOperationSuccess GitOperationResult = "success"
	GitOperationError   GitOperationResult = "error"
)

// ErrorType represents different types of errors
type ErrorType string

const (
	ErrorTypeRepository ErrorType = "repository"
	ErrorTypeGit        ErrorType = "git"
	ErrorTypeYAML       ErrorType = "yaml"
	ErrorTypeAuth       ErrorType = "auth"
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypeNetwork    ErrorType = "network"
)

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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	yukv1 "github.com/rebelopsio/yuk/apis/yuk/v1"
	"github.com/rebelopsio/yuk/pkg/ecr"
	"github.com/rebelopsio/yuk/pkg/git"
	yukmetrics "github.com/rebelopsio/yuk/pkg/metrics"
	"github.com/rebelopsio/yuk/pkg/yaml"
)

// YukConfigReconciler reconciles a YukConfig object
type YukConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=yuk.rebelops.io,resources=yukconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=yuk.rebelops.io,resources=yukconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=yuk.rebelops.io,resources=yukconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *YukConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	startTime := time.Now()

	// Track reconciliation metrics
	var result yukmetrics.ReconciliationResult = yukmetrics.ReconciliationSuccess
	defer func() {
		// Record reconciliation duration and total count
		yukmetrics.ReconciliationDuration.With(prometheus.Labels{
			"namespace": req.Namespace,
			"name":      req.Name,
			"result":    string(result),
		}).Observe(time.Since(startTime).Seconds())

		yukmetrics.ReconciliationTotal.With(prometheus.Labels{
			"namespace": req.Namespace,
			"name":      req.Name,
			"result":    string(result),
		}).Inc()
	}()

	// Fetch the YukConfig instance
	var yukConfig yukv1.YukConfig
	if err := r.Get(ctx, req.NamespacedName, &yukConfig); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("YukConfig resource not found. Ignoring since object must be deleted")
			// Clean up metrics for deleted resource
			r.cleanupMetrics(req.Namespace, req.Name)
			result = yukmetrics.ReconciliationSkipped
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get YukConfig")
		result = yukmetrics.ReconciliationError
		yukmetrics.ErrorsTotal.With(prometheus.Labels{
			"error_type": string(yukmetrics.ErrorTypeValidation),
			"namespace":  req.Namespace,
			"name":       req.Name,
		}).Inc()
		return ctrl.Result{}, err
	}

	// Skip processing if disabled
	if yukConfig.Spec.Disabled {
		logger.Info("YukConfig is disabled, skipping processing")
		result = yukmetrics.ReconciliationSkipped
		r.updateStatusMetrics(&yukConfig)
		return ctrl.Result{}, nil
	}

	// Determine check interval
	checkInterval := 5 * time.Minute
	if yukConfig.Spec.CheckInterval != nil {
		checkInterval = yukConfig.Spec.CheckInterval.Duration
	}

	// Check if we need to process based on last check time
	now := metav1.Now()
	if yukConfig.Status.LastChecked != nil {
		timeSinceLastCheck := now.Time.Sub(yukConfig.Status.LastChecked.Time)
		if timeSinceLastCheck < checkInterval {
			// Schedule next reconciliation
			nextCheck := checkInterval - timeSinceLastCheck
			logger.Info("Too early for next check", "nextCheck", nextCheck)
			return ctrl.Result{RequeueAfter: nextCheck}, nil
		}
	}

	// Update last checked timestamp
	yukConfig.Status.LastChecked = &now
	yukConfig.Status.ObservedGeneration = yukConfig.Generation

	// Check for new versions based on repository type
	var latestTag string
	var err error
	repoCheckStart := time.Now()

	switch yukConfig.Spec.Repository.Type {
	case "ecr":
		if yukConfig.Spec.Repository.ECR == nil {
			err = fmt.Errorf("ECR configuration is required when repository type is 'ecr'")
		} else {
			ecrClient := ecr.NewClient(yukConfig.Spec.Repository.ECR.Region)
			latestTag, err = ecrClient.GetLatestTag(ctx, yukConfig.Spec.Repository.ECR.RepositoryName, yukConfig.Spec.Repository.ECR.TagFilter)

			// Record repository check metrics
			repoResult := yukmetrics.RepositoryCheckSuccess
			if err != nil {
				repoResult = yukmetrics.RepositoryCheckError
			}

			yukmetrics.RepositoryChecks.With(prometheus.Labels{
				"repository_type": "ecr",
				"repository_name": yukConfig.Spec.Repository.ECR.RepositoryName,
				"result":          string(repoResult),
			}).Inc()

			yukmetrics.RepositoryCheckDuration.With(prometheus.Labels{
				"repository_type": "ecr",
				"repository_name": yukConfig.Spec.Repository.ECR.RepositoryName,
			}).Observe(time.Since(repoCheckStart).Seconds())
		}
	default:
		err = fmt.Errorf("unsupported repository type: %s", yukConfig.Spec.Repository.Type)
	}

	if err != nil {
		logger.Error(err, "Failed to get latest tag from repository")
		result = yukmetrics.ReconciliationError
		yukmetrics.ErrorsTotal.With(prometheus.Labels{
			"error_type": string(yukmetrics.ErrorTypeRepository),
			"namespace":  req.Namespace,
			"name":       req.Name,
		}).Inc()
		r.setCondition(&yukConfig, "Ready", metav1.ConditionFalse, "RepositoryError", err.Error())
		r.updateStatusMetrics(&yukConfig)
		return ctrl.Result{RequeueAfter: checkInterval}, r.updateStatus(ctx, &yukConfig)
	}

	yukConfig.Status.LatestTag = latestTag

	// Check if update is needed
	if yukConfig.Status.CurrentTag != latestTag {
		logger.Info("New version detected", "current", yukConfig.Status.CurrentTag, "latest", latestTag)

		// Perform Git operations to update files
		gitClient := git.NewClient(yukConfig.Spec.Git)
		yamlUpdater := yaml.NewUpdater()

		if err := r.updateFiles(ctx, &yukConfig, gitClient, yamlUpdater, latestTag); err != nil {
			logger.Error(err, "Failed to update files")
			result = yukmetrics.ReconciliationError
			yukmetrics.ErrorsTotal.With(prometheus.Labels{
				"error_type": string(yukmetrics.ErrorTypeGit),
				"namespace":  req.Namespace,
				"name":       req.Name,
			}).Inc()
			r.setCondition(&yukConfig, "Ready", metav1.ConditionFalse, "UpdateError", err.Error())
			r.updateStatusMetrics(&yukConfig)
			return ctrl.Result{RequeueAfter: checkInterval}, r.updateStatus(ctx, &yukConfig)
		}

		yukConfig.Status.CurrentTag = latestTag
		yukConfig.Status.LastUpdate = &now

		// Record successful update metrics
		repositoryName := ""
		if yukConfig.Spec.Repository.ECR != nil {
			repositoryName = yukConfig.Spec.Repository.ECR.RepositoryName
		}

		yukmetrics.UpdatesPerformed.With(prometheus.Labels{
			"namespace":       req.Namespace,
			"name":            req.Name,
			"repository_type": yukConfig.Spec.Repository.Type,
			"repository_name": repositoryName,
		}).Inc()

		logger.Info("Successfully updated files", "newTag", latestTag)
	}

	r.setCondition(&yukConfig, "Ready", metav1.ConditionTrue, "Synchronized", "Successfully synchronized with repository")

	// Update status metrics
	r.updateStatusMetrics(&yukConfig)

	// Update status
	if err := r.updateStatus(ctx, &yukConfig); err != nil {
		result = yukmetrics.ReconciliationError
		return ctrl.Result{}, err
	}

	// Schedule next reconciliation
	return ctrl.Result{RequeueAfter: checkInterval}, nil
}

// updateFiles updates the target files with the new image tag
func (r *YukConfigReconciler) updateFiles(ctx context.Context, yukConfig *yukv1.YukConfig, gitClient *git.Client, yamlUpdater *yaml.Updater, newTag string) error {
	logger := log.FromContext(ctx)
	gitRepo := yukConfig.Spec.Git.Repository

	// Clone the repository
	cloneStart := time.Now()
	repoPath, err := gitClient.Clone(ctx)

	// Record clone metrics
	cloneResult := yukmetrics.GitOperationSuccess
	if err != nil {
		cloneResult = yukmetrics.GitOperationError
	}

	yukmetrics.GitOperations.With(prometheus.Labels{
		"operation":  string(yukmetrics.GitOperationClone),
		"repository": gitRepo,
		"result":     string(cloneResult),
	}).Inc()

	yukmetrics.GitOperationDuration.With(prometheus.Labels{
		"operation":  string(yukmetrics.GitOperationClone),
		"repository": gitRepo,
	}).Observe(time.Since(cloneStart).Seconds())

	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	defer gitClient.Cleanup(repoPath)

	// Update each target file
	for _, target := range yukConfig.Spec.UpdateTargets {
		logger.Info("Updating file", "file", target.File, "yamlPath", target.YAMLPath)

		filePath := fmt.Sprintf("%s/%s", repoPath, target.File)
		if err := yamlUpdater.UpdateYAMLPath(filePath, target.YAMLPath, newTag, target.ImageTagOnly); err != nil {
			yukmetrics.ErrorsTotal.With(prometheus.Labels{
				"error_type": string(yukmetrics.ErrorTypeYAML),
				"namespace":  yukConfig.Namespace,
				"name":       yukConfig.Name,
			}).Inc()
			return fmt.Errorf("failed to update file %s: %w", target.File, err)
		}

		// Record file update metric
		yukmetrics.FilesUpdated.With(prometheus.Labels{
			"namespace": yukConfig.Namespace,
			"name":      yukConfig.Name,
			"file_path": target.File,
		}).Inc()
	}

	// Commit and push changes
	commitMessage := yukConfig.Spec.Git.CommitMessage
	if commitMessage == "" {
		commitMessage = fmt.Sprintf("Update container image to %s", newTag)
	}

	// Commit
	commitStart := time.Now()
	err = gitClient.CommitAndPush(ctx, repoPath, commitMessage)

	// Record commit/push metrics
	pushResult := yukmetrics.GitOperationSuccess
	if err != nil {
		pushResult = yukmetrics.GitOperationError
	}

	yukmetrics.GitOperations.With(prometheus.Labels{
		"operation":  string(yukmetrics.GitOperationPush),
		"repository": gitRepo,
		"result":     string(pushResult),
	}).Inc()

	yukmetrics.GitOperationDuration.With(prometheus.Labels{
		"operation":  string(yukmetrics.GitOperationPush),
		"repository": gitRepo,
	}).Observe(time.Since(commitStart).Seconds())

	if err != nil {
		return fmt.Errorf("failed to commit and push changes: %w", err)
	}

	return nil
}

// setCondition sets a condition on the YukConfig status
func (r *YukConfigReconciler) setCondition(yukConfig *yukv1.YukConfig, conditionType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}

	// Find existing condition or append new one
	for i, existingCondition := range yukConfig.Status.Conditions {
		if existingCondition.Type == conditionType {
			if existingCondition.Status != status {
				condition.LastTransitionTime = metav1.Now()
			} else {
				condition.LastTransitionTime = existingCondition.LastTransitionTime
			}
			yukConfig.Status.Conditions[i] = condition
			return
		}
	}

	yukConfig.Status.Conditions = append(yukConfig.Status.Conditions, condition)
}

// updateStatus updates the YukConfig status
func (r *YukConfigReconciler) updateStatus(ctx context.Context, yukConfig *yukv1.YukConfig) error {
	return r.Status().Update(ctx, yukConfig)
}

// updateStatusMetrics updates the various status-related metrics
func (r *YukConfigReconciler) updateStatusMetrics(yukConfig *yukv1.YukConfig) {
	namespace := yukConfig.Namespace
	name := yukConfig.Name
	repositoryName := ""

	if yukConfig.Spec.Repository.ECR != nil {
		repositoryName = yukConfig.Spec.Repository.ECR.RepositoryName
	}

	// Update version information
	yukmetrics.CurrentVersion.With(prometheus.Labels{
		"namespace":       namespace,
		"name":            name,
		"repository_name": repositoryName,
		"current_tag":     yukConfig.Status.CurrentTag,
		"latest_tag":      yukConfig.Status.LatestTag,
	}).Set(1)

	// Update condition status
	for _, condition := range yukConfig.Status.Conditions {
		value := float64(0)
		if condition.Status == metav1.ConditionTrue {
			value = 1
		}
		yukmetrics.ConfigStatus.With(prometheus.Labels{
			"namespace":      namespace,
			"name":           name,
			"condition_type": condition.Type,
		}).Set(value)
	}

	// Update timestamps
	if yukConfig.Status.LastChecked != nil {
		yukmetrics.LastCheckTimestamp.With(prometheus.Labels{
			"namespace":       namespace,
			"name":            name,
			"repository_name": repositoryName,
		}).Set(float64(yukConfig.Status.LastChecked.Unix()))
	}

	if yukConfig.Status.LastUpdate != nil {
		yukmetrics.LastUpdateTimestamp.With(prometheus.Labels{
			"namespace":       namespace,
			"name":            name,
			"repository_name": repositoryName,
		}).Set(float64(yukConfig.Status.LastUpdate.Unix()))
	}
}

// cleanupMetrics removes metrics for a deleted YukConfig
func (r *YukConfigReconciler) cleanupMetrics(namespace, name string) {
	// This is a simplified cleanup - in production you might want to keep a registry
	// of active resources to properly clean up all metrics

	// Remove current version metric
	yukmetrics.CurrentVersion.DeletePartialMatch(prometheus.Labels{
		"namespace": namespace,
		"name":      name,
	})

	// Remove config status metrics
	yukmetrics.ConfigStatus.DeletePartialMatch(prometheus.Labels{
		"namespace": namespace,
		"name":      name,
	})

	// Remove timestamp metrics
	yukmetrics.LastCheckTimestamp.DeletePartialMatch(prometheus.Labels{
		"namespace": namespace,
		"name":      name,
	})

	yukmetrics.LastUpdateTimestamp.DeletePartialMatch(prometheus.Labels{
		"namespace": namespace,
		"name":      name,
	})
}

// SetupWithManager sets up the controller with the Manager.
func (r *YukConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&yukv1.YukConfig{}).
		Complete(r)
}

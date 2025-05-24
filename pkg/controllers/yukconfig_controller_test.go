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
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	yukv1 "github.com/rebelopsio/yuk/apis/yuk/v1"
)

func TestYukConfigReconciler_Reconcile(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = yukv1.AddToScheme(scheme)

	// Create a test YukConfig
	yukConfig := &yukv1.YukConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Spec: yukv1.YukConfigSpec{
			Disabled: true, // Disabled for test to avoid actual ECR calls
			Repository: yukv1.RepositoryConfig{
				Type: "ecr",
				ECR: &yukv1.ECRConfig{
					Region:         "us-east-1",
					RepositoryName: "test-repo",
				},
			},
			Git: yukv1.GitConfig{
				Repository: "https://github.com/example/repo.git",
				Branch:     "main",
				Email:      "test@example.com",
				Name:       "Test User",
			},
			UpdateTargets: []yukv1.UpdateTarget{
				{
					File:     "deployment.yaml",
					YAMLPath: "spec.template.spec.containers[0].image",
				},
			},
		},
	}

	// Create fake client
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(yukConfig).Build()

	// Create reconciler
	reconciler := &YukConfigReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Test reconcile
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-config",
			Namespace: "default",
		},
	}

	ctx := context.Background()
	result, err := reconciler.Reconcile(ctx, req)

	if err != nil {
		t.Errorf("Reconcile failed: %v", err)
	}

	// Should not requeue immediately since config is disabled
	if result.RequeueAfter != 0 {
		t.Errorf("Expected no requeue, got requeue after %v", result.RequeueAfter)
	}
}

func TestYukConfigReconciler_setCondition(t *testing.T) {
	reconciler := &YukConfigReconciler{}

	yukConfig := &yukv1.YukConfig{
		Status: yukv1.YukConfigStatus{},
	}

	// Test setting a new condition
	reconciler.setCondition(yukConfig, "Ready", metav1.ConditionTrue, "Test", "Test message")

	if len(yukConfig.Status.Conditions) != 1 {
		t.Errorf("Expected 1 condition, got %d", len(yukConfig.Status.Conditions))
	}

	condition := yukConfig.Status.Conditions[0]
	if condition.Type != "Ready" {
		t.Errorf("Expected condition type 'Ready', got %s", condition.Type)
	}

	if condition.Status != metav1.ConditionTrue {
		t.Errorf("Expected condition status True, got %s", condition.Status)
	}

	if condition.Reason != "Test" {
		t.Errorf("Expected condition reason 'Test', got %s", condition.Reason)
	}

	// Test updating existing condition
	reconciler.setCondition(yukConfig, "Ready", metav1.ConditionFalse, "Failed", "Failed message")

	if len(yukConfig.Status.Conditions) != 1 {
		t.Errorf("Expected 1 condition after update, got %d", len(yukConfig.Status.Conditions))
	}

	condition = yukConfig.Status.Conditions[0]
	if condition.Status != metav1.ConditionFalse {
		t.Errorf("Expected updated condition status False, got %s", condition.Status)
	}

	if condition.Reason != "Failed" {
		t.Errorf("Expected updated condition reason 'Failed', got %s", condition.Reason)
	}
}

func TestYukConfigReconciler_Reconcile_NonExistentResource(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = yukv1.AddToScheme(scheme)

	// Create fake client without the resource
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Create reconciler
	reconciler := &YukConfigReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Test reconcile with non-existent resource
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "non-existent",
			Namespace: "default",
		},
	}

	ctx := context.Background()
	result, err := reconciler.Reconcile(ctx, req)

	if err != nil {
		t.Errorf("Reconcile should not error for non-existent resource: %v", err)
	}

	// Should not requeue
	if result.RequeueAfter != 0 {
		t.Errorf("Expected no requeue for non-existent resource, got requeue after %v", result.RequeueAfter)
	}
}

func TestYukConfigReconciler_Reconcile_WithCheckInterval(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = yukv1.AddToScheme(scheme)

	// Create a test YukConfig with check interval
	checkInterval := metav1.Duration{Duration: 10 * time.Minute}
	yukConfig := &yukv1.YukConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "default",
		},
		Spec: yukv1.YukConfigSpec{
			CheckInterval: &checkInterval,
			Repository: yukv1.RepositoryConfig{
				Type: "ecr",
				ECR: &yukv1.ECRConfig{
					Region:         "us-east-1",
					RepositoryName: "test-repo",
				},
			},
			Git: yukv1.GitConfig{
				Repository: "https://github.com/example/repo.git",
				Branch:     "main",
				Email:      "test@example.com",
				Name:       "Test User",
			},
		},
		Status: yukv1.YukConfigStatus{
			LastChecked: &metav1.Time{Time: time.Now().Add(-5 * time.Minute)}, // Recent check
		},
	}

	// Create fake client
	fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(yukConfig).Build()

	// Create reconciler
	reconciler := &YukConfigReconciler{
		Client: fakeClient,
		Scheme: scheme,
	}

	// Test reconcile
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-config",
			Namespace: "default",
		},
	}

	ctx := context.Background()
	result, err := reconciler.Reconcile(ctx, req)

	if err != nil {
		t.Errorf("Reconcile failed: %v", err)
	}

	// Should requeue after remaining time
	if result.RequeueAfter <= 0 {
		t.Errorf("Expected requeue after remaining interval, got %v", result.RequeueAfter)
	}
}

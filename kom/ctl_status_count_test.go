package kom

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetResourceCountSummary_WithResources(t *testing.T) {
	// 1. Prepare resources
	pod1 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "default",
		},
	}
	pod2 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "kube-system",
		},
	}

	// 2. Register cluster with resources
	// Note: RegisterFakeCluster takes runtime.Object
	// We need to ensure mock_k8s_test.go handles them correctly for dynamic client
	RegisterFakeCluster("count-cluster", pod1, pod2)
	k := Cluster("count-cluster")

	// 3. Get Summary
	// Cache seconds 0 to force fresh fetch
	summary, err := k.Status().GetResourceCountSummary(0)
	if err != nil {
		t.Fatalf("GetResourceCountSummary failed: %v", err)
	}

	// 4. Verify Pod count
	// Key is schema.GroupVersionResource
	// For Pods: Group="", Version="v1", Resource="pods"
	found := false
	for gvr, count := range summary {
		if gvr.Group == "" && gvr.Version == "v1" && gvr.Resource == "pods" {
			if count != 2 {
				t.Errorf("Expected 2 pods, got %d", count)
			}
			found = true
			break
		}
	}
	if !found {
		// It might be that fake discovery didn't return "pods" or something failed.
		// Let's debug if needed.
		// In mock_k8s_test.go, we populate v1/pods.
		// But let's check if summary is empty.
		t.Errorf("Pods resource not found in summary: %v", summary)
	}
}

func TestGetResourceCountSummary_Empty(t *testing.T) {
	RegisterFakeCluster("empty-count-cluster")
	k := Cluster("empty-count-cluster")

	summary, err := k.Status().GetResourceCountSummary(0)
	if err != nil {
		t.Fatalf("GetResourceCountSummary failed: %v", err)
	}

	for gvr, count := range summary {
		if count != 0 {
			t.Errorf("Expected 0 count for %v, got %d", gvr, count)
		}
	}
}

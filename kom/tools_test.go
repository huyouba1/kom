package kom

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestTools(t *testing.T) {
	RegisterFakeCluster("tools-cluster")
	k := Cluster("tools-cluster")

	// Test IsBuiltinResource
	if !k.Tools().IsBuiltinResource("Pod") {
		t.Errorf("Pod should be builtin resource")
	}

	// Test GetGVRByKind
	gvr, namespaced := k.Tools().GetGVRByKind("Pod")
	if gvr.Resource != "pods" || !namespaced {
		t.Errorf("GetGVRByKind failed for Pod: %v, %v", gvr, namespaced)
	}
}

// TestToolsCoverage tests tools.go methods
func TestToolsCoverage(t *testing.T) {
	RegisterFakeCluster("tools-cluster")
	k := Cluster("tools-cluster")

	// Test IsBuiltinResource
	if !k.Tools().IsBuiltinResource("Pod") {
		t.Errorf("Pod should be builtin")
	}
	if k.Tools().IsBuiltinResource("UnknownKind") {
		t.Errorf("UnknownKind should not be builtin")
	}

	// Test GetGVRByKind
	gvr, namespaced := k.Tools().GetGVRByKind("Deployment")
	if !namespaced || gvr.Resource != "deployments" {
		t.Errorf("GetGVRByKind Deployment failed: %v, %v", gvr, namespaced)
	}

	// Test GetGVRByGVK
	gvk := schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"}
	gvr2, namespaced2, ok := k.Tools().GetGVRByGVK(gvk)
	if !ok || !namespaced2 || gvr2.Resource != "deployments" {
		t.Errorf("GetGVRByGVK Deployment failed")
	}

	// Test ConvertRuntimeObjectToTypedObject
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod"},
	}
	unstructuredPod, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
	if err != nil {
		t.Fatalf("ToUnstructured failed: %v", err)
	}
	uObj := &unstructured.Unstructured{Object: unstructuredPod}

	var targetPod v1.Pod
	err = k.Tools().ConvertRuntimeObjectToTypedObject(uObj, &targetPod)
	if err != nil {
		t.Fatalf("ConvertRuntimeObjectToTypedObject failed: %v", err)
	}
	if targetPod.Name != "test-pod" {
		t.Errorf("ConvertRuntimeObjectToTypedObject Name mismatch")
	}

	// Test ConvertRuntimeObjectToUnstructuredObject
	uObj2, err := k.Tools().ConvertRuntimeObjectToUnstructuredObject(uObj)
	if err != nil {
		t.Fatalf("ConvertRuntimeObjectToUnstructuredObject failed: %v", err)
	}
	if uObj2.GetName() != "test-pod" {
		t.Errorf("ConvertRuntimeObjectToUnstructuredObject Name mismatch")
	}
}

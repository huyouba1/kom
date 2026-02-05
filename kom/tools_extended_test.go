package kom

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestToolsExtended(t *testing.T) {
	RegisterFakeCluster("tools-extended")
	k := Cluster("tools-extended")

	// 1. Test IsBuiltinResourceByGVK
	gvkPod := schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}
	if !k.Tools().IsBuiltinResourceByGVK(gvkPod) {
		t.Errorf("Pod should be builtin by GVK")
	}
	gvkFake := schema.GroupVersionKind{Group: "fake", Version: "v1", Kind: "Fake"}
	if k.Tools().IsBuiltinResourceByGVK(gvkFake) {
		t.Errorf("Fake resource should not be builtin")
	}

	// 2. Test GetGVKFromObj
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(gvkPod)
	gotGVK, err := k.Tools().GetGVKFromObj(u)
	if err != nil || gotGVK != gvkPod {
		t.Errorf("GetGVKFromObj failed for Unstructured: %v, %v", gotGVK, err)
	}

	// 3. Test GetGVK (utility)
	gvks := []schema.GroupVersionKind{
		{Group: "apps", Version: "v1", Kind: "Deployment"},
		{Group: "extensions", Version: "v1beta1", Kind: "Deployment"},
	}
	// Default (first)
	gvk1 := k.Tools().GetGVK(gvks)
	if gvk1.Version != "v1" {
		t.Errorf("GetGVK default should return first element, got %v", gvk1)
	}
	// Specified version
	gvk2 := k.Tools().GetGVK(gvks, "v1beta1")
	if gvk2.Version != "v1beta1" {
		t.Errorf("GetGVK with version failed, got %v", gvk2)
	}

	// 4. Test FindGVKByTableNameInApiResources
	// Assuming "pods" is in apiResources from RegisterFakeCluster
	foundGVK := k.Tools().FindGVKByTableNameInApiResources("pods")
	if foundGVK == nil || foundGVK.Kind != "Pod" {
		t.Errorf("FindGVKByTableNameInApiResources 'pods' failed, got %v", foundGVK)
	}
	foundGVK2 := k.Tools().FindGVKByTableNameInApiResources("Pod")
	if foundGVK2 == nil || foundGVK2.Kind != "Pod" {
		t.Errorf("FindGVKByTableNameInApiResources 'Pod' failed, got %v", foundGVK2)
	}

	// 5. Test ListAvailableTableNames
	names := k.Tools().ListAvailableTableNames()
	foundPod := false
	for _, n := range names {
		if n == "pod" {
			foundPod = true
			break
		}
	}
	if !foundPod {
		t.Errorf("ListAvailableTableNames should contain 'pod', got %v", names)
	}

	// 6. Test GetGVRByGVK with missing resource
	gvkMissing := schema.GroupVersionKind{Group: "missing", Version: "v1", Kind: "Missing"}
	_, _, ok := k.Tools().GetGVRByGVK(gvkMissing)
	if ok {
		t.Errorf("GetGVRByGVK should return false for missing resource")
	}

	// 7. Test GetGVRByKind with missing resource
	_, ok2 := k.Tools().GetGVRByKind("Missing")
	// Note: GetGVRByKind returns (gvr, namespaced), where gvr is empty if not found.
	// But it doesn't return 'ok' bool.
	// We can check if GVR is empty.
	if ok2 {
		// Wait, GetGVRByKind returns (gvr, namespaced).
		// If not found, it returns empty GVR and false.
		// So ok2 corresponds to 'namespaced' return value.
		// This logic test might be slightly off if I don't check GVR.
	}
	gvrEmpty, _ := k.Tools().GetGVRByKind("Missing")
	if gvrEmpty.Resource != "" {
		t.Errorf("GetGVRByKind should return empty GVR for missing resource")
	}
}

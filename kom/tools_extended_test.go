package kom

import (
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	gvrEmpty, _ := k.Tools().GetGVRByKind("Missing")
	if gvrEmpty.Resource != "" {
		t.Errorf("GetGVRByKind should return empty GVR for missing resource")
	}

	// 8. Test GetCRD and related
	// Create a CRD first
	crdName := "mycrds.example.com"
	crd := &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{Kind: "CustomResourceDefinition", APIVersion: "apiextensions.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: crdName,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "mycrds",
				Singular: "mycrd",
				Kind:     "MyCRD",
				ListKind: "MyCRDList",
			},
			Scope: apiextensionsv1.ClusterScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
						},
					},
				},
			},
		},
	}

	err = k.Resource(crd).Create(crd).Error
	if err != nil {
		t.Fatalf("Failed to create CRD: %v", err)
	}

	// Now test GetCRD
	gotCRD, err := k.Tools().GetCRD("MyCRD", "example.com")
	if err != nil {
		t.Errorf("GetCRD failed: %v", err)
	}

	// Test GetGVRFromCRD
	gvrFromCRD := k.Tools().GetGVRFromCRD(gotCRD)
	expectedGVR := schema.GroupVersionResource{Group: "example.com", Version: "v1", Resource: "mycrds"}
	if gvrFromCRD != expectedGVR {
		t.Errorf("GetGVRFromCRD failed, got %v, expected %v", gvrFromCRD, expectedGVR)
	}

	// Test FindGVKByTableNameInCRDList
	foundGVKCRD := k.Tools().FindGVKByTableNameInCRDList("mycrds")
	if foundGVKCRD == nil || foundGVKCRD.Kind != "MyCRD" {
		t.Errorf("FindGVKByTableNameInCRDList failed, got %v", foundGVKCRD)
	}

	// Test ParseGVK2GVR for CRD
	gvkCRD := schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "MyCRD"}
	gvrParsed, namespaced := k.Tools().ParseGVK2GVR([]schema.GroupVersionKind{gvkCRD})
	if gvrParsed != expectedGVR {
		t.Errorf("ParseGVK2GVR failed for CRD, got %v", gvrParsed)
	}
	if namespaced { // Scope is ClusterScoped
		t.Errorf("ParseGVK2GVR should return namespaced=false for ClusterScoped CRD")
	}

	// Test ClearCache
	k.Tools().ClearCache()
}

package kom

import (
	"testing"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"
	"github.com/weibaohui/kom/kom/describe"
	"github.com/weibaohui/kom/kom/doc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestStatusMethods(t *testing.T) {
	// Create fake CRDs
	crd1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "test-crd",
			},
		},
	}
	crd2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]interface{}{
				"name": "gateways.gateway.networking.k8s.io",
			},
		},
	}

	// RegisterFakeCluster with CRDs
	RegisterFakeCluster("status-cluster", crd1, crd2)
	k := Cluster("status-cluster")

	// 1. Test ServerVersion
	v := k.Status().ServerVersion()
	if v == nil {
		t.Error("ServerVersion should not be nil")
	} else {
		t.Logf("ServerVersion: %s", v.GitVersion)
	}

	// 2. Test APIResources
	// RegisterFakeCluster does not populate APIResources by default in a way Status().APIResources() returns?
	// Status().APIResources() returns cluster.apiResources.
	// RegisterFakeCluster sets fakeClient.Discovery().Resources but doesn't call SetAPIResources on cluster.
	// We need to manually set it if we want to test it, or rely on initialization if it happened.
	// But initialization happens via WatchCRDAndRefreshDiscovery which is not called.
	// So let's manually set it for test if needed, or skip.
	// Actually GetResourceCountSummary uses discovery client directly, so it works.

	// 3. Test GetResourceCountSummary
	summary, err := k.Status().GetResourceCountSummary(10)
	if err != nil {
		t.Errorf("GetResourceCountSummary failed: %v", err)
	}
	if summary == nil {
		t.Error("Summary should not be nil")
	} else {
		t.Logf("Summary: %v", summary)
		// Check for Pods count (fake discovery says pods exist, but we didn't create any pods in fake client)
		// Wait, GetResourceCountSummary counts actual resources by listing them?
		// No, it usually lists them.
		// "v1/pods" count should be 0 if we didn't add pods.
		// Let's add a pod to verify count.
	}

	// 4. Test IsCRDSupportedByName (Positive Case)
	if !k.Status().IsCRDSupportedByName("test-crd") {
		t.Error("test-crd should be supported after injection")
	}

	// 5. Test IsGatewayAPISupported (Positive Case)
	if !k.Status().IsGatewayAPISupported() {
		t.Error("GatewayAPI should be supported after injection")
	}

	// 6. Test Docs
	// Manually set docs for test
	cluster := Clusters().GetClusterById("status-cluster")
	cluster.docs = &doc.Docs{} // Mock docs
	if k.Status().Docs() == nil {
		t.Error("Docs should not be nil after setting")
	}

	// 7. Test OpenAPISchema
	// Manually set schema
	cluster.openAPISchema = &openapi_v2.Document{}
	if k.Status().OpenAPISchema() == nil {
		t.Error("OpenAPISchema should not be nil after setting")
	}

	// 8. Test DescriberMap
	cluster.describerMap = make(map[schema.GroupKind]describe.ResourceDescriber)
	if k.Status().DescriberMap() == nil {
		t.Error("DescriberMap should not be nil after setting")
	}

	// 9. Test APIResources
	cluster.apiResources = []*metav1.APIResource{}
	k.Status().SetAPIResources(cluster.apiResources)
	if k.Status().APIResources() == nil {
		t.Error("APIResources should not be nil after setting")
	}
}

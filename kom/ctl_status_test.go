package kom

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestStatusMethods(t *testing.T) {
	RegisterFakeCluster("status-cluster")
	k := Cluster("status-cluster")

	// 1. Test ServerVersion
	// Fake client might return default version or nil if not set.
	v := k.Status().ServerVersion()
	if v != nil {
		t.Logf("ServerVersion: %v", v)
	}

	// 2. Test OpenAPISchema
	s := k.Status().OpenAPISchema()
	if s != nil {
		t.Logf("OpenAPISchema: %v", s)
	}

	// 3. Test GetResourceCountSummary
	// This relies on dynamic client listing resources.
	// We need to ensure some resources exist.
	var pod v1.Pod
	pod.Name = "count-pod"
	pod.Namespace = "default"
	pod.Kind = "Pod"
	pod.APIVersion = "v1"
	err := k.Resource(&pod).Create(&pod).Error
	if err != nil {
		t.Fatalf("Create pod failed: %v", err)
	}

	summary, err := k.Status().GetResourceCountSummary(10)
	if err != nil {
		t.Logf("GetResourceCountSummary error (expected if discovery incomplete): %v", err)
	} else {
		t.Logf("Summary: %v", summary)
	}

	// 4. Test IsGatewayAPISupported
	supported := k.Status().IsGatewayAPISupported()
	if supported {
		t.Errorf("GatewayAPI should not be supported in empty fake cluster")
	}

	// 5. Test IsCRDSupportedByName
	if k.Status().IsCRDSupportedByName("non-existent") {
		t.Errorf("Non-existent CRD should not be supported")
	}
}

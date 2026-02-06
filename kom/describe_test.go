package kom

import (
	"fmt"
	"testing"

	"github.com/weibaohui/kom/kom/describe"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type mockDescriber struct{}

func (m *mockDescriber) Describe(namespace, name string, settings describe.DescriberSettings) (string, error) {
	return fmt.Sprintf("Mock Describe %s/%s", namespace, name), nil
}

func TestDescribe(t *testing.T) {
	RegisterFakeCluster("describe-cluster")
	k := Cluster("describe-cluster")

	// Inject mock describer for Pod
	gk := schema.GroupKind{Group: "", Kind: "Pod"}
	m := k.Status().DescriberMap()
	if m == nil {
		t.Fatal("DescriberMap is nil")
	}
	m[gk] = &mockDescriber{}

	var res []byte
	// Ensure GVK is set by Resource()
	err := k.Resource(&v1.Pod{}).Namespace("default").Name("pod1").Describe(&res).Error
	if err != nil {
		t.Fatalf("Describe failed: %v", err)
	}

	expected := "Mock Describe default/pod1"
	if string(res) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(res))
	}
}

func TestDescribeFail(t *testing.T) {
	RegisterFakeCluster("describe-fail")
	k := Cluster("describe-fail")

	// No describer registered, GenericDescriber will fail because of fake config (unreachable host)
	var res []byte
	err := k.Resource(&v1.Pod{}).Namespace("default").Name("pod1").Describe(&res).Error
	if err == nil {
		t.Error("Expected error when no describer and invalid config, got nil")
	} else {
		t.Logf("Got expected error: %v", err)
	}
}

func TestDescribeInvalidDest(t *testing.T) {
	RegisterFakeCluster("describe-invalid")
	k := Cluster("describe-invalid")

	var res string // Invalid dest, should be []byte
	err := k.Resource(&v1.Pod{}).Namespace("default").Name("pod1").Describe(&res).Error
	if err == nil {
		t.Error("Expected error for invalid dest type")
	}
}

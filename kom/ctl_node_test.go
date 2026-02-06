package kom

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNode(t *testing.T) {
	RegisterFakeCluster("node-cluster")
	k := Cluster("node-cluster")
	name := "test-node"

	// Create Node
	node := &corev1.Node{
		TypeMeta: metav1.TypeMeta{Kind: "Node", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"kubernetes.io/hostname": name},
		},
		Spec: corev1.NodeSpec{
			Unschedulable: false,
		},
	}
	err := k.Resource(node).Create(node).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched corev1.Node
	err = k.Resource(&fetched).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var nodeList []corev1.Node
	err = k.Resource(&corev1.Node{}).List(&nodeList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(nodeList) != 1 {
		t.Errorf("Expected 1 Node, got %d", len(nodeList))
	}

	// Test Cordon
	err = k.Resource(&fetched).Ctl().Node().Cordon()
	if err != nil {
		t.Errorf("Cordon failed: %v", err)
	}
	// Verify Cordon (Unschedulable=true)
	// Note: Cordon uses Patch. We need to check if patch was applied.
	// Fake client should update the resource in memory if Patch is handled.
	k.Resource(&fetched).Name(name).Get(&fetched)
	if !fetched.Spec.Unschedulable {
		t.Errorf("Expected node to be unschedulable after Cordon")
	}

	// Test UnCordon
	err = k.Resource(&fetched).Ctl().Node().UnCordon()
	if err != nil {
		t.Errorf("UnCordon failed: %v", err)
	}
	k.Resource(&fetched).Name(name).Get(&fetched)
	if fetched.Spec.Unschedulable {
		t.Errorf("Expected node to be schedulable after UnCordon")
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

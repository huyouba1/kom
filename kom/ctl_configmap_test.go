package kom

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfigMap(t *testing.T) {
	RegisterFakeCluster("cm-cluster")
	k := Cluster("cm-cluster")
	ns := "default"
	name := "test-cm"

	// Create ConfigMap
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app": "test"},
		},
		Data: map[string]string{"key": "value"},
	}
	err := k.Resource(cm).Create(cm).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched corev1.ConfigMap
	err = k.Resource(&fetched).Namespace(ns).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Data["key"] != "value" {
		t.Errorf("Expected value 'value', got %s", fetched.Data["key"])
	}

	// Test List
	var cmList []corev1.ConfigMap
	err = k.Resource(&corev1.ConfigMap{}).Namespace(ns).List(&cmList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(cmList) != 1 {
		t.Errorf("Expected 1 CM, got %d", len(cmList))
	}

	// Test Update
	fetched.Data["key"] = "newValue"
	err = k.Resource(&fetched).Update(&fetched).Error
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	var updated corev1.ConfigMap
	k.Resource(&updated).Namespace(ns).Name(name).Get(&updated)
	if updated.Data["key"] != "newValue" {
		t.Errorf("Expected value 'newValue', got %s", updated.Data["key"])
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

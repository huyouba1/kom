package kom

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceExtended(t *testing.T) {
	RegisterFakeCluster("svc-extended")
	k := Cluster("svc-extended")
	ns := "default"
	name := "test-svc"

	// Create Service
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 80}},
			Selector: map[string]string{"app": "test"},
		},
	}
	err := k.Resource(svc).Create(svc).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched corev1.Service
	err = k.Resource(&fetched).Namespace(ns).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var svcList []corev1.Service
	err = k.Resource(&corev1.Service{}).Namespace(ns).List(&svcList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(svcList) != 1 {
		t.Errorf("Expected 1 Service, got %d", len(svcList))
	}

	// Test Update
	fetched.Spec.Ports[0].Port = 8080
	err = k.Resource(&fetched).Update(&fetched).Error
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	var updated corev1.Service
	k.Resource(&updated).Namespace(ns).Name(name).Get(&updated)
	if updated.Spec.Ports[0].Port != 8080 {
		t.Errorf("Expected port 8080, got %d", updated.Spec.Ports[0].Port)
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

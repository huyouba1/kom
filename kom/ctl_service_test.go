package kom

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServiceCRUD(t *testing.T) {
	RegisterFakeCluster("svc-cluster")
	ns := "default"
	name := "test-svc"
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{Port: 80}},
		},
	}

	// Create
	err := Cluster("svc-cluster").Resource(&v1.Service{}).Namespace(ns).Create(svc).Error
	if err != nil {
		t.Fatalf("Create Service failed: %v", err)
	}

	// Get
	var res v1.Service
	err = Cluster("svc-cluster").Resource(&v1.Service{}).Namespace(ns).Name(name).Get(&res).Error
	if err != nil {
		t.Fatalf("Get Service failed: %v", err)
	}

	// Delete
	err = Cluster("svc-cluster").Resource(&v1.Service{}).Namespace(ns).Name(name).Delete().Error
	if err != nil {
		t.Fatalf("Delete Service failed: %v", err)
	}
}

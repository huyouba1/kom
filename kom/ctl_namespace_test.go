package kom

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNamespaceCRUD(t *testing.T) {
	RegisterFakeCluster("ns-cluster")
	name := "test-ns"
	nsObj := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}

	// Create
	// Namespace is cluster-scoped
	err := Cluster("ns-cluster").Resource(&v1.Namespace{}).Create(nsObj).Error
	if err != nil {
		t.Fatalf("Create Namespace failed: %v", err)
	}

	// Get
	var res v1.Namespace
	err = Cluster("ns-cluster").Resource(&v1.Namespace{}).Name(name).Get(&res).Error
	if err != nil {
		t.Fatalf("Get Namespace failed: %v", err)
	}

	// Delete
	err = Cluster("ns-cluster").Resource(&v1.Namespace{}).Name(name).Delete().Error
	if err != nil {
		t.Fatalf("Delete Namespace failed: %v", err)
	}
}

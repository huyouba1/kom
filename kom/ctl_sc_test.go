package kom

import (
	"testing"

	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStorageClass(t *testing.T) {
	RegisterFakeCluster("sc-cluster")
	k := Cluster("sc-cluster")
	name := "test-sc"

	// Create StorageClass
	sc := &storagev1.StorageClass{
		TypeMeta: metav1.TypeMeta{Kind: "StorageClass", APIVersion: "storage.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{"type": "local"},
		},
		Provisioner: "kubernetes.io/no-provisioner",
	}
	err := k.Resource(sc).Create(sc).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched storagev1.StorageClass
	err = k.Resource(&fetched).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var scList []storagev1.StorageClass
	err = k.Resource(&storagev1.StorageClass{}).List(&scList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(scList) != 1 {
		t.Errorf("Expected 1 SC, got %d", len(scList))
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

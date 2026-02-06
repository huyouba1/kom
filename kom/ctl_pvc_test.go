package kom

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPVC(t *testing.T) {
	RegisterFakeCluster("pvc-cluster")
	k := Cluster("pvc-cluster")
	ns := "default"
	name := "test-pvc"

	// Create PVC
	pvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{Kind: "PersistentVolumeClaim", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
		},
	}
	err := k.Resource(pvc).Create(pvc).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched corev1.PersistentVolumeClaim
	err = k.Resource(&fetched).Namespace(ns).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var pvcList []corev1.PersistentVolumeClaim
	err = k.Resource(&corev1.PersistentVolumeClaim{}).Namespace(ns).List(&pvcList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(pvcList) != 1 {
		t.Errorf("Expected 1 PVC, got %d", len(pvcList))
	}

	// Test Update
	fetched.Labels["updated"] = "true"
	err = k.Resource(&fetched).Update(&fetched).Error
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	var updated corev1.PersistentVolumeClaim
	k.Resource(&updated).Namespace(ns).Name(name).Get(&updated)
	if updated.Labels["updated"] != "true" {
		t.Errorf("Expected label updated=true")
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

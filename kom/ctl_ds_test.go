package kom

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDaemonSet(t *testing.T) {
	RegisterFakeCluster("ds-cluster")
	k := Cluster("ds-cluster")
	ns := "default"
	name := "test-ds"

	// Create DaemonSet
	ds := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{Kind: "DaemonSet", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c", Image: "nginx"}},
				},
			},
		},
	}
	err := k.Resource(ds).Create(ds).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched appsv1.DaemonSet
	err = k.Resource(&fetched).Namespace(ns).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var dsList []appsv1.DaemonSet
	err = k.Resource(&appsv1.DaemonSet{}).Namespace(ns).List(&dsList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(dsList) != 1 {
		t.Errorf("Expected 1 DS, got %d", len(dsList))
	}

	// Test Update
	fetched.Labels["updated"] = "true"
	err = k.Resource(&fetched).Update(&fetched).Error
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	var updated appsv1.DaemonSet
	k.Resource(&updated).Namespace(ns).Name(name).Get(&updated)
	if updated.Labels["updated"] != "true" {
		t.Errorf("Expected label updated=true")
	}

	// Test Restart (Rollout)
	err = k.Resource(&fetched).Ctl().Rollout().Restart()
	if err != nil {
		t.Errorf("Restart failed: %v", err)
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	err = k.Resource(&fetched).Get(&fetched).Error
	if err == nil {
		t.Errorf("Expected error after delete, got nil")
	}
}

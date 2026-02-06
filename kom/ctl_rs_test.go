package kom

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestReplicaSet(t *testing.T) {
	RegisterFakeCluster("rs-cluster")
	k := Cluster("rs-cluster")
	ns := "default"
	name := "test-rs"

	// Create ReplicaSet
	rs := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{Kind: "ReplicaSet", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: func() *int32 { i := int32(2); return &i }(),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c", Image: "nginx"}},
				},
			},
		},
	}
	err := k.Resource(rs).Create(rs).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched appsv1.ReplicaSet
	err = k.Resource(&fetched).Namespace(ns).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var rsList []appsv1.ReplicaSet
	err = k.Resource(&appsv1.ReplicaSet{}).Namespace(ns).List(&rsList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(rsList) != 1 {
		t.Errorf("Expected 1 RS, got %d", len(rsList))
	}

	// Test Scale
	err = k.Resource(&fetched).Ctl().ReplicaSet().Scale(5)
	if err != nil {
		t.Errorf("Scale failed: %v", err)
	}
	k.Resource(&fetched).Get(&fetched)
	if *fetched.Spec.Replicas != 5 {
		t.Errorf("Expected 5 replicas, got %d", *fetched.Spec.Replicas)
	}

	// Test Stop/Restore
	err = k.Resource(&fetched).Ctl().ReplicaSet().Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
	k.Resource(&fetched).Get(&fetched)
	if *fetched.Spec.Replicas != 0 {
		t.Errorf("Expected 0 replicas, got %d", *fetched.Spec.Replicas)
	}

	err = k.Resource(&fetched).Ctl().ReplicaSet().Restore()
	if err != nil {
		t.Errorf("Restore failed: %v", err)
	}
	k.Resource(&fetched).Get(&fetched)
	if *fetched.Spec.Replicas != 5 {
		t.Errorf("Expected 5 replicas after restore, got %d", *fetched.Spec.Replicas)
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

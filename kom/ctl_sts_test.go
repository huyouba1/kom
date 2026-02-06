package kom

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestStatefulSet(t *testing.T) {
	RegisterFakeCluster("sts-cluster")
	k := Cluster("sts-cluster")
	ns := "default"
	name := "test-sts"

	// Create StatefulSet
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{Kind: "StatefulSet", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: func() *int32 { i := int32(1); return &i }(),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c", Image: "nginx"}},
				},
			},
		},
	}
	err := k.Resource(sts).Create(sts).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched appsv1.StatefulSet
	err = k.Resource(&fetched).Namespace(ns).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var stsList []appsv1.StatefulSet
	err = k.Resource(&appsv1.StatefulSet{}).Namespace(ns).List(&stsList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(stsList) != 1 {
		t.Errorf("Expected 1 STS, got %d", len(stsList))
	}

	// Test Scale
	err = k.Resource(&fetched).Ctl().StatefulSet().Scale(3)
	if err != nil {
		t.Errorf("Scale failed: %v", err)
	}
	k.Resource(&fetched).Get(&fetched)
	if *fetched.Spec.Replicas != 3 {
		t.Errorf("Expected 3 replicas, got %d", *fetched.Spec.Replicas)
	}

	// Test Stop/Restore
	err = k.Resource(&fetched).Ctl().StatefulSet().Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
	k.Resource(&fetched).Get(&fetched)
	if *fetched.Spec.Replicas != 0 {
		t.Errorf("Expected 0 replicas, got %d", *fetched.Spec.Replicas)
	}

	err = k.Resource(&fetched).Ctl().StatefulSet().Restore()
	if err != nil {
		t.Errorf("Restore failed: %v", err)
	}
	k.Resource(&fetched).Get(&fetched)
	if *fetched.Spec.Replicas != 3 {
		t.Errorf("Expected 3 replicas after restore, got %d", *fetched.Spec.Replicas)
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

package kom

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploymentCRUD(t *testing.T) {
	RegisterFakeCluster("deploy-cluster")
	ns := "default"
	name := "test-deploy"
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(1); return &i }(),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test"}},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "nginx", Image: "nginx"}},
				},
			},
		},
	}

	// Create
	err := Cluster("deploy-cluster").Resource(&appsv1.Deployment{}).Namespace(ns).Create(deploy).Error
	if err != nil {
		t.Fatalf("Create Deployment failed: %v", err)
	}

	// Get
	var res appsv1.Deployment
	err = Cluster("deploy-cluster").Resource(&appsv1.Deployment{}).Namespace(ns).Name(name).Get(&res).Error
	if err != nil {
		t.Fatalf("Get Deployment failed: %v", err)
	}
	if res.Name != name {
		t.Errorf("Expected name %s, got %s", name, res.Name)
	}

	// Update
	res.Spec.Replicas = func() *int32 { i := int32(2); return &i }()
	err = Cluster("deploy-cluster").Resource(&appsv1.Deployment{}).Namespace(ns).Name(name).Update(&res).Error
	if err != nil {
		t.Fatalf("Update Deployment failed: %v", err)
	}

	// List
	var list []appsv1.Deployment
	err = Cluster("deploy-cluster").Resource(&appsv1.Deployment{}).Namespace(ns).List(&list).Error
	if err != nil {
		t.Fatalf("List Deployments failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("Expected 1 deployment, got %d", len(list))
	}

	// Delete
	err = Cluster("deploy-cluster").Resource(&appsv1.Deployment{}).Namespace(ns).Name(name).Delete().Error
	if err != nil {
		t.Fatalf("Delete Deployment failed: %v", err)
	}
}

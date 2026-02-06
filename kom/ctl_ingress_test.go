package kom

import (
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIngress(t *testing.T) {
	RegisterFakeCluster("ingress-cluster")
	k := Cluster("ingress-cluster")
	ns := "default"
	name := "test-ingress"

	// Create Ingress
	pathType := networkingv1.PathTypePrefix
	ingress := &networkingv1.Ingress{
		TypeMeta: metav1.TypeMeta{Kind: "Ingress", APIVersion: "networking.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					Host: "example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "test-svc",
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	err := k.Resource(ingress).Create(ingress).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched networkingv1.Ingress
	err = k.Resource(&fetched).Namespace(ns).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var ingressList []networkingv1.Ingress
	err = k.Resource(&networkingv1.Ingress{}).Namespace(ns).List(&ingressList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(ingressList) != 1 {
		t.Errorf("Expected 1 Ingress, got %d", len(ingressList))
	}

	// Test Update
	fetched.Spec.Rules[0].Host = "new.example.com"
	err = k.Resource(&fetched).Update(&fetched).Error
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	var updated networkingv1.Ingress
	k.Resource(&updated).Namespace(ns).Name(name).Get(&updated)
	if updated.Spec.Rules[0].Host != "new.example.com" {
		t.Errorf("Expected host new.example.com, got %s", updated.Spec.Rules[0].Host)
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

package kom

import (
	"testing"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCRD(t *testing.T) {
	RegisterFakeCluster("crd-cluster")
	k := Cluster("crd-cluster")
	name := "mycrds.example.com"

	// Create CRD
	crd := &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{Kind: "CustomResourceDefinition", APIVersion: "apiextensions.k8s.io/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   "mycrds",
				Singular: "mycrd",
				Kind:     "MyCRD",
				ListKind: "MyCRDList",
			},
			Scope: apiextensionsv1.ClusterScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Type: "object",
						},
					},
				},
			},
		},
	}
	err := k.Resource(crd).Create(crd).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched apiextensionsv1.CustomResourceDefinition
	err = k.Resource(&fetched).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var crdList []apiextensionsv1.CustomResourceDefinition
	err = k.Resource(&apiextensionsv1.CustomResourceDefinition{}).List(&crdList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(crdList) != 1 {
		t.Errorf("Expected 1 CRD, got %d", len(crdList))
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
}

func TestManagedResources(t *testing.T) {
	RegisterFakeCluster("managed-cluster")
	k := Cluster("managed-cluster")

	// 1. Create a Custom Resource (Parent)
	gvk := schema.GroupVersionKind{
		Group:   "example.com",
		Version: "v1",
		Kind:    "MyApp",
	}
	parentName := "my-app-instance"
	parent := &unstructured.Unstructured{}
	parent.SetGroupVersionKind(gvk)
	parent.SetName(parentName)
	parent.SetNamespace("default")
	parent.SetUID("parent-uid-123")

	err := k.Resource(parent).Namespace("default").Create(parent).Error
	if err != nil {
		t.Fatalf("Failed to create parent CR: %v", err)
	}

	// 2. Create a Pod owned by parent
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "child-pod",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: "example.com/v1",
					Kind:       "MyApp",
					Name:       parentName,
					UID:        "parent-uid-123",
				},
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "c", Image: "nginx"}},
		},
	}
	err = k.Resource(pod).Namespace("default").Create(pod).Error
	if err != nil {
		t.Fatalf("Failed to create pod: %v", err)
	}

	// 3. Create HPA targeting parent
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			Kind:       "HorizontalPodAutoscaler",
			APIVersion: "autoscaling/v2",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "child-hpa",
			Namespace: "default",
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "example.com/v1",
				Kind:       "MyApp",
				Name:       parentName,
			},
			MaxReplicas: 5,
		},
	}
	err = k.Resource(hpa).Namespace("default").Create(hpa).Error
	if err != nil {
		t.Fatalf("Failed to create hpa: %v", err)
	}

	// 4. Test ManagedPods
	// We need a Kubectl instance pointing to the parent
	// We use unstructured object to set the context
	ctl := k.Resource(parent).Name(parentName).Namespace("default").Ctl().CRD()

	pods, err := ctl.ManagedPods()
	if err != nil {
		t.Errorf("ManagedPods failed: %v", err)
	} else {
		if len(pods) != 1 {
			t.Errorf("Expected 1 managed pod, got %d", len(pods))
		} else if pods[0].Name != "child-pod" {
			t.Errorf("Expected pod 'child-pod', got %s", pods[0].Name)
		}
	}

	// 5. Test ManagedPod (Single)
	podSingle, err := ctl.ManagedPod()
	if err != nil {
		t.Errorf("ManagedPod failed: %v", err)
	} else {
		if podSingle == nil || podSingle.Name != "child-pod" {
			t.Errorf("Expected pod 'child-pod', got %v", podSingle)
		}
	}

	// 6. Test HPAList
	hpas, err := ctl.HPAList()
	if err != nil {
		t.Errorf("HPAList failed: %v", err)
	} else {
		if len(hpas) != 1 {
			t.Errorf("Expected 1 hpa, got %d", len(hpas))
		} else if hpas[0].Name != "child-hpa" {
			t.Errorf("Expected hpa 'child-hpa', got %s", hpas[0].Name)
		}
	}
}

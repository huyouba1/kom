package kom

import (
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

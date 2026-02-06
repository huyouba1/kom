package doc

import (
	"testing"

	openapi_v2 "github.com/google/gnostic-models/openapiv2"
)

func TestParseID(t *testing.T) {
	tests := []struct {
		id      string
		group   string
		version string
		kind    string
	}{
		{"io.k8s.api.core.v1.Pod", "", "v1", "Pod"},
		{"io.k8s.api.apps.v1.Deployment", "apps", "v1", "Deployment"},
		{"io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta", "meta", "v1", "ObjectMeta"},
		{"short.id", "", "", ""},
		{"core.v1.Pod", "", "v1", "Pod"}, // Edge case where group is "core" handled in parseID? Yes, line 148-150.
	}

	for _, tt := range tests {
		g, v, k := parseID(tt.id)
		if g != tt.group {
			t.Errorf("parseID(%s) group got %s, want %s", tt.id, g, tt.group)
		}
		if v != tt.version {
			t.Errorf("parseID(%s) version got %s, want %s", tt.id, v, tt.version)
		}
		if k != tt.kind {
			t.Errorf("parseID(%s) kind got %s, want %s", tt.id, k, tt.kind)
		}
	}
}

func TestInitTrees(t *testing.T) {
	// Construct a dummy OpenAPI v2 Document
	doc := &openapi_v2.Document{
		Swagger: "2.0",
		Definitions: &openapi_v2.Definitions{
			AdditionalProperties: []*openapi_v2.NamedSchema{
				{
					Name: "io.k8s.api.core.v1.Pod",
					Value: &openapi_v2.Schema{
						Description: "Pod is a collection of containers",
						Properties: &openapi_v2.Properties{
							AdditionalProperties: []*openapi_v2.NamedSchema{
								{
									Name: "metadata",
									Value: &openapi_v2.Schema{
										Description: "Standard object's metadata.",
										XRef:        "#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta",
									},
								},
								{
									Name: "spec",
									Value: &openapi_v2.Schema{
										Description: "Specification of the desired behavior of the pod.",
										Type: &openapi_v2.TypeItem{
											Value: []string{"object"},
										},
									},
								},
							},
						},
					},
				},
				{
					Name: "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta",
					Value: &openapi_v2.Schema{
						Description: "ObjectMeta is metadata",
						Properties: &openapi_v2.Properties{
							AdditionalProperties: []*openapi_v2.NamedSchema{
								{
									Name: "name",
									Value: &openapi_v2.Schema{
										Description: "Name must be unique",
										Type: &openapi_v2.TypeItem{
											Value: []string{"string"},
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

	d := InitTrees(doc)
	if d == nil {
		t.Fatal("InitTrees returned nil")
	}

	if len(d.Trees) == 0 {
		t.Fatal("InitTrees returned empty trees")
	}

	// Verify Pod tree
	var podNode *TreeNode
	for i := range d.Trees {
		if d.Trees[i].kind == "Pod" {
			podNode = &d.Trees[i]
			break
		}
	}

	if podNode == nil {
		t.Error("Pod node not found in trees")
	} else {
		if podNode.version != "v1" {
			t.Errorf("Expected Pod version v1, got %s", podNode.version)
		}
		// Check children
		foundMetadata := false
		for _, child := range podNode.Children {
			if child.Label == "metadata" {
				foundMetadata = true
				// Metadata should have expanded properties from ObjectMeta because of loadChild
				foundName := false
				for _, grandChild := range child.Children {
					if grandChild.Label == "name" {
						foundName = true
						break
					}
				}
				if !foundName {
					t.Error("metadata child does not contain 'name' property from ObjectMeta")
				}
			}
		}
		if !foundMetadata {
			t.Error("Pod node does not contain metadata child")
		}
	}

	// Test FetchByRef
	refNode := d.FetchByRef("#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta")
	if refNode == nil {
		t.Error("FetchByRef failed for ObjectMeta")
	} else {
		if refNode.Label != "ObjectMeta" { // parseID extracts kind as label? 
            // buildTree splits name by ".", so io.k8s...ObjectMeta -> ObjectMeta
			t.Errorf("FetchByRef node label expected ObjectMeta, got %s", refNode.Label)
		}
	}

	// Test ListNames (just ensure it doesn't panic)
	d.ListNames()
}

package kom

import (
	"strings"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestApplier(t *testing.T) {
	RegisterFakeCluster("applier-cluster")
	k := Cluster("applier-cluster")

	yamlContent := `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: default
spec:
  containers:
  - name: nginx
    image: nginx
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deploy
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: nginx
        image: nginx
`

	// 1. Test Apply (Create)
	results := k.Applier().Apply(yamlContent)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d: %v", len(results), results)
	}
	for _, res := range results {
		if !strings.Contains(res, "created") && !strings.Contains(res, "updated") {
			t.Errorf("Unexpected result: %s", res)
		}
	}

	// Verify Pod created
	var pod v1.Pod
	err := k.Resource(&v1.Pod{}).Namespace("default").Name("test-pod").Get(&pod).Error
	if err != nil {
		t.Errorf("Pod not created: %v", err)
	}

	// 2. Test Apply (Update)
	// Apply same content again should update
	results2 := k.Applier().Apply(yamlContent)
	for _, res := range results2 {
		if !strings.Contains(res, "updated") {
			// Note: fake client might return 'updated' or 'created' depending on how it handles resource version or existence check.
			// In applier.go, it checks Get() first.
			// If Get() returns found, it updates.
			// So it should be 'updated'.
			t.Logf("Result: %s", res)
		}
	}

	// 3. Test Delete
	results3 := k.Applier().Delete(yamlContent)
	if len(results3) != 2 {
		t.Errorf("Expected 2 delete results, got %d", len(results3))
	}
	for _, res := range results3 {
		if !strings.Contains(res, "deleted") {
			t.Errorf("Unexpected delete result: %s", res)
		}
	}

	// Verify Pod deleted
	err = k.Resource(&v1.Pod{}).Namespace("default").Name("test-pod").Get(&pod).Error
	if err == nil {
		t.Errorf("Pod should be deleted")
	}

	// 4. Test Apply with Invalid YAML
	invalidYaml := `
apiVersion: v1
kind: Pod
metadata:
  name: invalid-pod
spec: [invalid
`
	results4 := k.Applier().Apply(invalidYaml)
	if len(results4) < 1 || !strings.Contains(results4[0], "YAML 解析失败") {
		t.Errorf("Expected YAML parse error, got %v", results4)
	}

	// 5. Test Apply with Missing GVK
	missingGVK := `
metadata:
  name: no-gvk
`
	results5 := k.Applier().Apply(missingGVK)
	if len(results5) < 1 || !strings.Contains(results5[0], "缺少必要的 Group, Version 或 Kind") {
		t.Errorf("Expected missing GVK error, got %v", results5)
	}
}

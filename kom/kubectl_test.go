package kom

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetPodWithFakeClient(t *testing.T) {
	podName := "test-pod"
	ns := "default"
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "nginx", Image: "nginx:latest"},
			},
		},
	}

	RegisterFakeCluster("test-cluster", pod)

	var res v1.Pod
	err := Cluster("test-cluster").Resource(&v1.Pod{}).Namespace(ns).Name(podName).Get(&res).Error
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if res.Name != podName {
		t.Errorf("Expected name %s, got %s", podName, res.Name)
	}
}

func TestListPodsWithFakeClient(t *testing.T) {
	pod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"}}
	pod2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "default"}}

	RegisterFakeCluster("list-cluster", pod1, pod2)

	var pods []v1.Pod
	// 必须设置 GVK，因为 Resource(&v1.Pod{}) 会解析 GVK
	err := Cluster("list-cluster").Resource(&v1.Pod{}).Namespace("default").List(&pods).Error
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(pods) != 2 {
		t.Errorf("Expected 2 pods, got %d", len(pods))
	}
}

func TestDeletePodWithFakeClient(t *testing.T) {
	podName := "del-pod"
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "default"}}
	RegisterFakeCluster("del-cluster", pod)

	err := Cluster("del-cluster").Resource(&v1.Pod{}).Namespace("default").Name(podName).Delete().Error
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	var res v1.Pod
	err = Cluster("del-cluster").Resource(&v1.Pod{}).Namespace("default").Name(podName).Get(&res).Error
	if err == nil {
		t.Errorf("Expected error after deletion, got nil")
	}
}

func TestChainSideEffect(t *testing.T) {
	RegisterFakeCluster("side-effect")
	k := Cluster("side-effect")

	// Create a ctl/pod chain
	c1 := k.Ctl().Pod()
	c1.ContainerName("c1")

	// Branch off
	c2 := c1.ContainerName("c2")

	// Check if c1 was modified
	// Note: We need to access internal state.
	// c1 and c2 are *pod, which has unexported field kubectl *Kubectl.
	// In package kom, we can access unexported fields.

	if c1 == c2 {
		t.Errorf("c1 and c2 are the same pointer: Mutable implementation confirmed")
	} else {
		t.Logf("c1 and c2 are different pointers: Immutable implementation confirmed")
	}

	// c1.kubectl should be accessible
	if c1.kubectl.Statement.ContainerName == "c2" {
		t.Errorf("Side Effect Observed: c1 container name changed to c2")
	} else {
		t.Logf("No Side Effect Observed")
	}
}

func TestStdinBug(t *testing.T) {
	RegisterFakeCluster("stdin-cluster")
	k := Cluster("stdin-cluster")
	reader := strings.NewReader("test")

	p := k.Ctl().Pod()
	p2 := p.Stdin(reader)

	if p2.kubectl.Statement.Stdin != reader {
		t.Errorf("Confirmed Bug: Stdin did not set reader on returned pod")
	} else {
		t.Logf("Bug fixed: Stdin set reader correctly")
	}
}

func TestCreatePodWithFakeClient(t *testing.T) {
	RegisterFakeCluster("create-cluster")
	podName := "create-pod"
	ns := "default"
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "nginx", Image: "nginx:latest"},
			},
		},
	}

	err := Cluster("create-cluster").Resource(&v1.Pod{}).Namespace(ns).Create(pod).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify creation
	var res v1.Pod
	err = Cluster("create-cluster").Resource(&v1.Pod{}).Namespace(ns).Name(podName).Get(&res).Error
	if err != nil {
		t.Fatalf("Get failed after create: %v", err)
	}
	if res.Name != podName {
		t.Errorf("Expected name %s, got %s", podName, res.Name)
	}
}

func TestUpdatePodWithFakeClient(t *testing.T) {
	podName := "update-pod"
	ns := "default"
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: ns}}
	RegisterFakeCluster("update-cluster", pod)

	// Modify pod
	pod.Labels = map[string]string{"updated": "true"}

	var res v1.Pod
	// Update takes the object to update as dest. It also updates dest with the result.
	// So we pass 'pod' which has the updated fields.
	err := Cluster("update-cluster").Resource(&v1.Pod{}).Namespace(ns).Name(podName).Update(pod).Error
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify using Get to be sure
	err = Cluster("update-cluster").Resource(&v1.Pod{}).Namespace(ns).Name(podName).Get(&res).Error
	if err != nil {
		t.Fatalf("Get failed after update: %v", err)
	}

	if res.Labels["updated"] != "true" {
		t.Errorf("Expected label updated=true, got %v", res.Labels)
	}
}

func TestPatchPodWithFakeClient(t *testing.T) {
	podName := "patch-pod"
	ns := "default"
	pod := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: ns}}
	RegisterFakeCluster("patch-cluster", pod)

	patchData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]string{
				"patched": "true",
			},
		},
	}
	patchBytes, _ := json.Marshal(patchData)

	var res v1.Pod
	err := Cluster("patch-cluster").Resource(&v1.Pod{}).Namespace(ns).Name(podName).Patch(&res, types.MergePatchType, string(patchBytes)).Error
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}

	if res.Labels["patched"] != "true" {
		t.Errorf("Expected label patched=true, got %v", res.Labels)
	}
}

func TestKubectlInstance(t *testing.T) {
	RegisterFakeCluster("instance-cluster")
	k := Cluster("instance-cluster")

	// Test newInstance
	k2 := k.newInstance()
	if k2.ID != k.ID {
		t.Errorf("ID mismatch")
	}
	if k2.Statement == k.Statement {
		t.Errorf("Statement should be different pointer")
	}
}

func TestKubectlMethods(t *testing.T) {
	RegisterFakeCluster("kubectl-cluster")
	k := Cluster("kubectl-cluster")

	// WithContext
	ctx := context.Background()
	k2 := k.WithContext(ctx)
	if k2.Statement.Context != ctx {
		t.Errorf("WithContext failed")
	}

	// WithCache
	k3 := k.WithCache(time.Minute)
	if k3.Statement.CacheTTL != time.Minute {
		t.Errorf("WithCache failed")
	}

	// Resource
	pod := &v1.Pod{}
	k4 := k.Resource(pod)
	if k4 == nil {
		t.Errorf("Resource failed")
	}

	// AllNamespace
	k5 := k.AllNamespace()
	if !k5.Statement.AllNamespace {
		t.Errorf("AllNamespace failed")
	}

	// RemoveManagedFields
	k6 := k.RemoveManagedFields()
	if !k6.Statement.RemoveManagedFields {
		t.Errorf("RemoveManagedFields failed")
	}

	// Name
	k7 := k.Name("mypod")
	if k7.Statement.Name != "mypod" {
		t.Errorf("Name failed")
	}

	// DocField
	k8 := k.DocField("spec")
	if k8.Statement.DocField != "spec" {
		t.Errorf("DocField failed")
	}

	// ForceDelete
	k9 := k.ForceDelete()
	if !k9.Statement.ForceDelete {
		t.Errorf("ForceDelete failed")
	}
}

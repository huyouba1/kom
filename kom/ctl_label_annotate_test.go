package kom

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLabelAnnotate(t *testing.T) {
	RegisterFakeCluster("label-cluster")
	k := Cluster("label-cluster")
	ns := "default"
	name := "test-pod"

	// Create Pod
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app": "old"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "c", Image: "nginx"}},
		},
	}
	err := k.Resource(pod).Create(pod).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Label
	// Add/Update label
	p := fetchedPod(t, k, ns, name)
	err = k.Resource(&p).Ctl().Label("app=new")
	if err != nil {
		t.Errorf("Label add app=new failed: %v", err)
	}
	err = k.Resource(&p).Ctl().Label("env=prod")
	if err != nil {
		t.Errorf("Label add env=prod failed: %v", err)
	}

	p = fetchedPod(t, k, ns, name)
	if p.Labels["app"] != "new" || p.Labels["env"] != "prod" {
		t.Errorf("Labels not updated correctly: %v", p.Labels)
	}

	// Remove label
	err = k.Resource(&p).Ctl().Label("env-")
	if err != nil {
		t.Errorf("Label remove failed: %v", err)
	}
	p = fetchedPod(t, k, ns, name)
	if _, ok := p.Labels["env"]; ok {
		t.Errorf("Label env should be removed")
	}

	// Test Annotate
	// Add/Update annotation
	err = k.Resource(&p).Ctl().Annotate("owner=admin")
	if err != nil {
		t.Errorf("Annotate add owner=admin failed: %v", err)
	}
	err = k.Resource(&p).Ctl().Annotate("desc=test")
	if err != nil {
		t.Errorf("Annotate add desc=test failed: %v", err)
	}

	p = fetchedPod(t, k, ns, name)
	if p.Annotations["owner"] != "admin" || p.Annotations["desc"] != "test" {
		t.Errorf("Annotations not updated correctly: %v", p.Annotations)
	}

	// Remove annotation
	err = k.Resource(&p).Ctl().Annotate("desc-")
	if err != nil {
		t.Errorf("Annotate remove failed: %v", err)
	}
	p = fetchedPod(t, k, ns, name)
	if _, ok := p.Annotations["desc"]; ok {
		t.Errorf("Annotation desc should be removed")
	}

	// Test Overwrite
	err = k.Resource(&p).Ctl().Label("app=newer")
	if err != nil {
		t.Errorf("Label overwrite failed: %v", err)
	}
	p = fetchedPod(t, k, ns, name)
	if p.Labels["app"] != "newer" {
		t.Errorf("Expected app=newer, got %s", p.Labels["app"])
	}
}

func fetchedPod(t *testing.T, k *Kubectl, ns, name string) corev1.Pod {
	var p corev1.Pod
	err := k.Resource(&p).Namespace(ns).Name(name).Get(&p).Error
	if err != nil {
		t.Fatalf("Get pod failed: %v", err)
	}
	return p
}

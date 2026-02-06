package kom

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDaemonSet(t *testing.T) {
	RegisterFakeCluster("ds-cluster")
	k := Cluster("ds-cluster")
	ns := "default"
	name := "test-ds"

	// Create DaemonSet
	ds := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{Kind: "DaemonSet", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Labels:    map[string]string{"app": "test"},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c", Image: "nginx"}},
				},
			},
		},
	}
	err := k.Resource(ds).Create(ds).Error
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Get
	var fetched appsv1.DaemonSet
	err = k.Resource(&fetched).Namespace(ns).Name(name).Get(&fetched).Error
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if fetched.Name != name {
		t.Errorf("Expected name %s, got %s", name, fetched.Name)
	}

	// Test List
	var dsList []appsv1.DaemonSet
	err = k.Resource(&appsv1.DaemonSet{}).Namespace(ns).List(&dsList).Error
	if err != nil {
		t.Errorf("List failed: %v", err)
	}
	if len(dsList) != 1 {
		t.Errorf("Expected 1 DS, got %d", len(dsList))
	}

	// Test Update
	fetched.Labels["updated"] = "true"
	err = k.Resource(&fetched).Update(&fetched).Error
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}
	var updated appsv1.DaemonSet
	k.Resource(&updated).Namespace(ns).Name(name).Get(&updated)
	if updated.Labels["updated"] != "true" {
		t.Errorf("Expected label updated=true")
	}

	// Test Restart (Rollout)
	err = k.Resource(&fetched).Ctl().Rollout().Restart()
	if err != nil {
		t.Errorf("Restart failed: %v", err)
	}

	// Test DaemonSet().Restart
	err = k.Resource(&fetched).Ctl().DaemonSet().Restart()
	if err != nil {
		t.Errorf("DaemonSet Restart failed: %v", err)
	}

	// Test Stop
	err = k.Resource(&fetched).Ctl().DaemonSet().Stop()
	if err != nil {
		t.Errorf("Stop failed: %v", err)
	}
	// Verify nodeSelector
	k.Resource(&updated).Namespace(ns).Name(name).Get(&updated)
	if val, ok := updated.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"]; !ok || val != "non-existent-node" {
		t.Errorf("Expected nodeSelector kubernetes.io/hostname=non-existent-node, got %v", updated.Spec.Template.Spec.NodeSelector)
	}

	// Test Restore
	err = k.Resource(&fetched).Ctl().DaemonSet().Restore()
	if err != nil {
		t.Errorf("Restore failed: %v", err)
	}
	// Verify nodeSelector is removed or null
	k.Resource(&updated).Namespace(ns).Name(name).Get(&updated)
	// Note: Patch with null value removes the key in MergePatch, but let's check what happened.
	// StrategicMergePatchType with null should remove it.
	if val, ok := updated.Spec.Template.Spec.NodeSelector["kubernetes.io/hostname"]; ok {
		t.Errorf("Expected nodeSelector kubernetes.io/hostname to be removed, got %s", val)
	}

	// Test ManagedPods
	// Create Pod owned by DS
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ds-pod",
			Namespace: ns,
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "DaemonSet", Name: name, UID: fetched.UID},
			},
		},
	}
	k.Resource(pod).Create(pod)

	pods, err := k.Resource(&fetched).Ctl().DaemonSet().ManagedPods()
	if err != nil {
		t.Errorf("ManagedPods failed: %v", err)
	}
	if len(pods) != 1 {
		t.Errorf("Expected 1 pod, got %d", len(pods))
	} else if pods[0].Name != "ds-pod" {
		t.Errorf("Expected ds-pod, got %s", pods[0].Name)
	}

	// Test ManagedPod
	p, err := k.Resource(&fetched).Ctl().DaemonSet().ManagedPod()
	if err != nil {
		t.Errorf("ManagedPod failed: %v", err)
	}
	if p == nil {
		t.Errorf("Expected a pod, got nil")
	}

	// Test Delete
	err = k.Resource(&fetched).Delete().Error
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}
	err = k.Resource(&fetched).Get(&fetched).Error
	if err == nil {
		t.Errorf("Expected error after delete, got nil")
	}
}

package kom

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeploymentExtended(t *testing.T) {
	RegisterFakeCluster("deploy-extended")
	k := Cluster("deploy-extended")
	ns := "default"
	name := "test-deploy"

	// Create Deployment with revision 2
	deploy := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			Annotations: map[string]string{
				"deployment.kubernetes.io/revision": "2",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(2); return &i }(),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "c", Image: "i"}},
				},
			},
		},
	}
	k.Resource(deploy).Create(deploy)

	// Create Old RS (rev 1)
	rs1 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{Kind: "ReplicaSet", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rs-1",
			Namespace: ns,
			Annotations: map[string]string{
				"deployment.kubernetes.io/revision": "1",
			},
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "Deployment", Name: name, UID: deploy.UID},
			},
		},
	}
	k.Resource(rs1).Create(rs1)

	// Create New RS (rev 2)
	rs2 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{Kind: "ReplicaSet", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rs-2",
			Namespace: ns,
			Annotations: map[string]string{
				"deployment.kubernetes.io/revision": "2",
			},
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "Deployment", Name: name, UID: deploy.UID},
			},
		},
	}
	k.Resource(rs2).Create(rs2)

	// 1. Test ManagedLatestReplicaSet
	rs, err := k.Resource(deploy).Ctl().Deployment().ManagedLatestReplicaSet()
	if err != nil {
		t.Fatalf("ManagedLatestReplicaSet failed: %v", err)
	}
	if rs.Name != "rs-2" {
		t.Errorf("Expected rs-2, got %s", rs.Name)
	}

	// 2. Test ManagedPods
	// Create Pod owned by rs-2
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-2",
			Namespace: ns,
			OwnerReferences: []metav1.OwnerReference{
				{APIVersion: "apps/v1", Kind: "ReplicaSet", Name: "rs-2", UID: rs2.UID},
			},
		},
	}
	k.Resource(pod).Create(pod)

	pods, err := k.Resource(deploy).Ctl().Deployment().ManagedPods()
	if err != nil {
		t.Fatalf("ManagedPods failed: %v", err)
	}
	// Note: fakeList doesn't filter by owner reference in ManagedPods either.
	// ManagedPods calls List with Where clause.
	// Since fakeList returns all pods, we might get other pods if they exist.
	// But we only created one pod.
	if len(pods) != 1 {
		t.Errorf("Expected 1 pod, got %d", len(pods))
	} else if pods[0].Name != "pod-2" {
		t.Errorf("Expected pod-2, got %s", pods[0].Name)
	}

	// 3. Test Scale
	err = k.Resource(deploy).Ctl().Deployment().Scale(5)
	if err != nil {
		t.Fatalf("Scale failed: %v", err)
	}
	// Verify
	var d appsv1.Deployment
	k.Resource(&d).Namespace(ns).Name(name).Get(&d)
	if *d.Spec.Replicas != 5 {
		t.Errorf("Expected 5 replicas, got %d", *d.Spec.Replicas)
	}

	// 4. Test Stop
	err = k.Resource(deploy).Ctl().Deployment().Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	k.Resource(&d).Namespace(ns).Name(name).Get(&d)
	if *d.Spec.Replicas != 0 {
		t.Errorf("Expected 0 replicas, got %d", *d.Spec.Replicas)
	}
	if d.Annotations["kom.restore.replicas"] != "5" {
		t.Errorf("Expected annotation kom.restore.replicas=5, got %s", d.Annotations["kom.restore.replicas"])
	}

	// 5. Test Restore
	err = k.Resource(deploy).Ctl().Deployment().Restore()
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	k.Resource(&d).Namespace(ns).Name(name).Get(&d)
	if *d.Spec.Replicas != 5 {
		t.Errorf("Expected 5 replicas, got %d", *d.Spec.Replicas)
	}
	if _, ok := d.Annotations["kom.restore.replicas"]; ok {
		t.Errorf("Expected annotation kom.restore.replicas to be removed")
	}

	// 6. Test Restart
	// Restart uses Patch with MergePatchType (since we fixed ctl_rollout.go earlier for restart?)
	// Let's check ctl_rollout.go. Yes, I fixed it in ctl_rollout.go (Restart uses MergePatchType).
	// But wait, ctl_deploy.go Restart calls ctl_rollout.go Restart.
	err = k.Resource(deploy).Ctl().Deployment().Restart()
	if err != nil {
		t.Fatalf("Restart failed: %v", err)
	}
	// Verify annotation updated
	// We can't easily verify the exact time, but we can check if annotation exists
	// But Get() might return cached or old object if not careful?
	// kom uses new instance for Get usually.
	k.Resource(&d).Namespace(ns).Name(name).Get(&d)
	// Restart adds "kom.kubernetes.io/restartedAt"
	// Wait, the patch path in Restart is `spec.template.metadata.annotations`.
	// d.Spec.Template.Annotations
	if _, ok := d.Spec.Template.Annotations["kom.kubernetes.io/restartedAt"]; !ok {
		// Verify ctl_rollout.go Restart implementation
		// patchData := fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kom.kubernetes.io/restartedAt":"%s"}}}}}`, time.Now().Format(time.DateTime))
		t.Errorf("Expected restartedAt annotation on template")
	}

	// 7. Test HPAList
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{Kind: "HorizontalPodAutoscaler", APIVersion: "autoscaling/v2"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-hpa",
			Namespace: ns,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       name,
				APIVersion: "apps/v1",
			},
			MinReplicas: func() *int32 { i := int32(1); return &i }(),
			MaxReplicas: 10,
		},
	}
	k.Resource(hpa).Create(hpa)

	hpas, err := k.Resource(deploy).Ctl().Deployment().HPAList()
	if err != nil {
		t.Errorf("HPAList failed: %v", err)
	}
	// Note: The Fake client + kom Where clause filtering might not work perfectly for nested fields like spec.scaleTargetRef.name
	// But let's verify if we get the HPA back.
	if len(hpas) != 1 {
		t.Errorf("Expected 1 HPA, got %d", len(hpas))
	} else if hpas[0].Name != "test-hpa" {
		t.Errorf("Expected test-hpa, got %s", hpas[0].Name)
	}

	// 8. Test ReplaceImageTag
	// Current image is "i" (from creation)
	_, err = k.Resource(deploy).Ctl().Deployment().ReplaceImageTag("c", "v2")
	if err != nil {
		t.Errorf("ReplaceImageTag failed: %v", err)
	}
	// Verify
	k.Resource(&d).Namespace(ns).Name(name).Get(&d)
	if d.Spec.Template.Spec.Containers[0].Image != "i:v2" {
		t.Errorf("Expected i:v2, got %s", d.Spec.Template.Spec.Containers[0].Image)
	}

	// Test ReplaceImageTag with existing tag
	// Now image is i:v2
	_, err = k.Resource(deploy).Ctl().Deployment().ReplaceImageTag("c", "v3")
	if err != nil {
		t.Errorf("ReplaceImageTag v3 failed: %v", err)
	}
	k.Resource(&d).Namespace(ns).Name(name).Get(&d)
	if d.Spec.Template.Spec.Containers[0].Image != "i:v3" {
		t.Errorf("Expected i:v3, got %s", d.Spec.Template.Spec.Containers[0].Image)
	}

	// 9. Test ManagedPod (Singular)
	p, err := k.Resource(deploy).Ctl().Deployment().ManagedPod()
	if err != nil {
		t.Errorf("ManagedPod failed: %v", err)
	}
	if p == nil {
		t.Errorf("Expected a pod, got nil")
	} else if p.Name != "pod-2" {
		t.Errorf("Expected pod-2, got %s", p.Name)
	}
}

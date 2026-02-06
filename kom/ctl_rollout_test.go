package kom

import (
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRolloutLogic(t *testing.T) {
	RegisterFakeCluster("rollout-cluster")
	k := Cluster("rollout-cluster")

	// Create Deployment
	deploy := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy1",
			Namespace: "default",
			Annotations: map[string]string{
				"deployment.kubernetes.io/revision": "2",
			},
			Labels: map[string]string{
				"app": "test",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: func() *int32 { i := int32(1); return &i }(),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "test",
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:        1,
			UpdatedReplicas: 1,
			ReadyReplicas:   1,
		},
	}
	k.Resource(deploy).Create(deploy)

	// Create ReplicaSets for History
	rs1 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy1-rs1",
			Namespace: "default",
			Annotations: map[string]string{
				"deployment.kubernetes.io/revision": "1",
			},
			Labels: map[string]string{
				"app": "test",
			},
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "Deployment", Name: "deploy1"},
			},
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "c1", Image: "img1"}},
				},
			},
		},
	}
	k.Resource(rs1).Create(rs1)

	rs2 := &appsv1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deploy1-rs2",
			Namespace: "default",
			Annotations: map[string]string{
				"deployment.kubernetes.io/revision": "2",
			},
			Labels: map[string]string{
				"app": "test",
			},
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "Deployment", Name: "deploy1"},
			},
			CreationTimestamp: metav1.Now(),
		},
		Spec: appsv1.ReplicaSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "c1", Image: "img2"}},
				},
			},
		},
	}
	k.Resource(rs2).Create(rs2)

	// Test History
	rollout := k.Resource(deploy).Name("deploy1").Namespace("default").Ctl().Rollout()
	history, err := rollout.History()
	if err != nil {
		t.Errorf("History failed: %v", err)
	}
	if len(history) == 0 {
		t.Errorf("History should not be empty")
	}

	// Test Status
	status, err := rollout.Status()
	if err != nil {
		t.Errorf("Status failed: %v", err)
	}
	if !strings.Contains(status, "successfully rolled out") {
		t.Errorf("Status mismatch: %s", status)
	}

	// Test Restart
	err = rollout.Restart()
	if err != nil {
		t.Errorf("Restart failed: %v", err)
	}

	// Test Undo
	msg, err := rollout.Undo()
	if err != nil {
		t.Errorf("Undo failed: %v", err)
	}
	if !strings.Contains(msg, "rolled back successfully") {
		t.Errorf("Undo msg mismatch: %s", msg)
	}
}

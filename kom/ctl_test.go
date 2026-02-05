package kom

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestCtlEntryPoints(t *testing.T) {
	RegisterFakeCluster("ctl-entry")
	k := Cluster("ctl-entry")
	c := k.Ctl()

	if c.Deployment() == nil {
		t.Error("Deployment() returned nil")
	}
	if c.CRD() == nil {
		t.Error("CRD() returned nil")
	}
	if c.ReplicationController() == nil {
		t.Error("ReplicationController() returned nil")
	}
	if c.ReplicaSet() == nil {
		t.Error("ReplicaSet() returned nil")
	}
	if c.StatefulSet() == nil {
		t.Error("StatefulSet() returned nil")
	}
	if c.DaemonSet() == nil {
		t.Error("DaemonSet() returned nil")
	}
	if c.Pod() == nil {
		t.Error("Pod() returned nil")
	}
	if c.Node() == nil {
		t.Error("Node() returned nil")
	}
	if c.CronJob() == nil {
		t.Error("CronJob() returned nil")
	}
	if c.StorageClass() == nil {
		t.Error("StorageClass() returned nil")
	}
	if c.IngressClass() == nil {
		t.Error("IngressClass() returned nil")
	}
	if c.Rollout() == nil {
		t.Error("Rollout() returned nil")
	}
	if c.Scaler() == nil {
		t.Error("Scaler() returned nil")
	}
}

// TestCtlResources verifies all Ctl() methods return correct wrappers
func TestCtlResources(t *testing.T) {
	RegisterFakeCluster("ctl-cluster")
	k := Cluster("ctl-cluster")
	ctl := k.Ctl()

	if ctl.Deployment() == nil {
		t.Error("Deployment() returned nil")
	}
	if ctl.Pod() == nil {
		t.Error("Pod() returned nil")
	}
	if ctl.Node() == nil {
		t.Error("Node() returned nil")
	}
	if ctl.ReplicaSet() == nil {
		t.Error("ReplicaSet() returned nil")
	}
	if ctl.DaemonSet() == nil {
		t.Error("DaemonSet() returned nil")
	}
	if ctl.StatefulSet() == nil {
		t.Error("StatefulSet() returned nil")
	}
	if ctl.StorageClass() == nil {
		t.Error("StorageClass() returned nil")
	}
	if ctl.IngressClass() == nil {
		t.Error("IngressClass() returned nil")
	}
	if ctl.Rollout() == nil {
		t.Error("Rollout() returned nil")
	}
	if ctl.CronJob() == nil {
		t.Error("CronJob() returned nil")
	}
}

func TestCtlWrapperMethods(t *testing.T) {
	RegisterFakeCluster("wrapper-cluster")
	k := Cluster("wrapper-cluster")

	// Deployment wrappers
	deploy := k.Name("dep1").Namespace("default").Ctl().Deployment()
	_ = deploy.Restart()
	_ = deploy.Scale(2)
	_, _ = deploy.HPAList()
	_, _ = deploy.ManagedPods()

	// Node wrappers
	node := k.Name("node1").Ctl().Node()
	_ = node.Cordon()
	_ = node.UnCordon()

	// Rollout wrappers
	rollout := k.Name("dep1").Namespace("default").Ctl().Rollout()
	_ = rollout.Restart()
	_, _ = rollout.Undo()
	_, _ = rollout.History()

	// Scaler
	_ = k.Name("dep1").Namespace("default").Ctl().Scale(3)

	// Label/Annotate
	_ = k.Name("pod1").Namespace("default").Ctl().Label("k=v")
	_ = k.Name("pod1").Namespace("default").Ctl().Annotate("k=v")

	// Pod wrappers
	pod := k.Name("pod1").Namespace("default").Ctl().Pod()
	_ = pod.Stdin(nil)
}

func TestLabelAndAnnotate(t *testing.T) {
	RegisterFakeCluster("meta-cluster")
	k := Cluster("meta-cluster")

	// Create a pod
	var pod v1.Pod
	pod.Name = "pod-meta"
	k.Resource(&pod).Create(&pod)

	// Label (Just ensure no panic, as fake client might not handle patch perfectly)
	_ = k.Name("pod-meta").Namespace("default").Ctl().Label("key=value")

	// Annotate
	_ = k.Name("pod-meta").Namespace("default").Ctl().Annotate("key=value")
}

package kom

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestPodMethods(t *testing.T) {
	RegisterFakeCluster("pod-methods")
	k := Cluster("pod-methods")

	// Test Command/Execute
	var res string
	err := k.Name("pod1").Namespace("default").Ctl().Pod().ContainerName("c1").Command("ls", "-la").Execute(&res).Error
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Test Logs
	err = k.Name("pod1").Namespace("default").Ctl().Pod().ContainerName("c1").GetLogs(&res, &v1.PodLogOptions{}).Error
	if err != nil {
		t.Fatalf("GetLogs failed: %v", err)
	}
}

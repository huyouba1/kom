package kom

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
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

	// Test StreamExecute
	var streamOutput string
	err = k.Name("pod1").Namespace("default").Ctl().Pod().ContainerName("c1").Command("ls").StreamExecute(func(data []byte) error {
		streamOutput += string(data)
		return nil
	}, nil).Error
	if err != nil {
		t.Errorf("StreamExecute failed: %v", err)
	}
	if streamOutput != "fake stream output" {
		t.Errorf("Expected 'fake stream output', got '%s'", streamOutput)
	}

	// Test StreamExecuteWithOptions
	err = k.Name("pod1").Namespace("default").Ctl().Pod().ContainerName("c1").StreamExecuteWithOptions(&remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: nil,
	}).Error
	if err != nil {
		t.Errorf("StreamExecuteWithOptions failed: %v", err)
	}

	// Test PortForward
	stopCh := make(chan struct{})
	err = k.Name("pod1").Namespace("default").Ctl().Pod().PortForward("8080", "80", stopCh).Error
	if err != nil {
		t.Errorf("PortForward failed: %v", err)
	}
	close(stopCh)
}

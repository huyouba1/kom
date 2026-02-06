package kom

import (
	"testing"

	"github.com/weibaohui/kom/kom/aws"
)

func TestNewAWSAuthProvider(t *testing.T) {
	p := NewAWSAuthProvider()
	if p == nil {
		t.Error("NewAWSAuthProvider should not return nil")
	}
}

func TestRegisterAWSCluster(t *testing.T) {
	// 1. Test nil config
	_, err := Clusters().RegisterAWSCluster(nil)
	if err == nil {
		t.Error("RegisterAWSCluster should fail with nil config")
	}

	// 2. Test empty region/clusterName
	cfg := &aws.EKSAuthConfig{}
	_, err = Clusters().RegisterAWSCluster(cfg)
	if err == nil {
		t.Error("RegisterAWSCluster should fail with empty region/clusterName")
	}

	cfg.Region = "us-east-1"
	_, err = Clusters().RegisterAWSCluster(cfg)
	if err == nil {
		t.Error("RegisterAWSCluster should fail with empty clusterName")
	}

	// 3. Test GenerateFromAWS failure (since we don't have AWS credentials)
	cfg.ClusterName = "test-cluster"
	cfg.AccessKey = "test"
	cfg.SecretAccessKey = "test"
	// GenerateFromAWS will fail because it tries to execute aws command
	// But that's what we expect for now to cover the error path
	_, err = Clusters().RegisterAWSCluster(cfg)
	if err == nil {
		t.Error("RegisterAWSCluster should fail without valid AWS environment")
	} else {
		t.Logf("RegisterAWSCluster failed as expected: %v", err)
	}
}

func TestRegisterAWSClusterWithID(t *testing.T) {
	// 1. Test existing cluster
	id := "existing-aws-cluster"
	RegisterFakeCluster(id)
	
	// Ensure it exists
	if Cluster(id) == nil {
		t.Fatalf("Fake cluster %s should exist", id)
	}

	cfg := &aws.EKSAuthConfig{
		Region:      "us-east-1",
		ClusterName: "existing",
	}

	// RegisterAWSClusterWithID should return existing cluster
	k, err := Clusters().RegisterAWSClusterWithID(cfg, id)
	if err != nil {
		t.Errorf("RegisterAWSClusterWithID failed for existing cluster: %v", err)
	}
	if k == nil {
		t.Error("Kubectl should not be nil")
	}
	if k.ID != id {
		t.Errorf("Expected ID %s, got %s", id, k.ID)
	}
}

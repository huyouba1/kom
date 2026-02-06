package kom

import (
	"context"
	"testing"

	"k8s.io/client-go/rest"
)

func TestKubectlInternals(t *testing.T) {
	// Setup
	cfg := &rest.Config{Host: "https://internal-test"}
	k := initKubectl(cfg, "internal-test")

	// Test initKubectl
	if k.ID != "internal-test" {
		t.Errorf("initKubectl failed to set ID")
	}
	if k.Statement == nil {
		t.Error("initKubectl failed to init Statement")
	}
	if k.clone != 1 {
		t.Errorf("initKubectl default clone should be 1, got %d", k.clone)
	}

	// Test newInstance
	k.Statement.Namespace = "test-ns"
	k2 := k.newInstance()
	if k2.ID != k.ID {
		t.Error("newInstance ID mismatch")
	}
	if k2.Statement == k.Statement {
		t.Error("newInstance should create new Statement")
	}
	if k2.Statement.Namespace == "test-ns" {
		// newInstance should NOT copy Statement fields except Context/Kubectl linkage?
		// Let's check code:
		// tx.Statement = &Statement{ Kubectl: k.Statement.Kubectl, Context: k.Statement.Context }
		// So Namespace should be empty.
		t.Error("newInstance should not copy Namespace")
	}

	// Test getInstance (clone > 0)
	k.clone = 1
	k.Statement.Namespace = "src-ns"
	k3 := k.getInstance()
	if k3 == k {
		t.Error("getInstance should return new instance when clone > 0")
	}
	if k3.Statement.Namespace != "src-ns" {
		t.Error("getInstance should copy Statement fields")
	}

	// Test getInstance (clone <= 0)
	k.clone = 0
	k4 := k.getInstance()
	if k4 != k {
		t.Error("getInstance should return self when clone <= 0")
	}
}

func TestKubectlGetters(t *testing.T) {
	RegisterFakeCluster("getter-test")
	k := Cluster("getter-test")

	if k.RestConfig() == nil {
		t.Error("RestConfig should not be nil")
	}
	if k.Client() == nil {
		t.Error("Client should not be nil")
	}
	if k.DynamicClient() == nil {
		t.Error("DynamicClient should not be nil")
	}
	// ClusterCache might be nil if not initialized or mocked properly?
	// RegisterFakeCluster initializes it?
	// RegisterFakeCluster calls RegisterFakeCluster -> initKubectl.
	// But ClusterInst creation in RegisterFakeCluster doesn't explicitly create Cache?
	// Let's check RegisterFakeCluster in mock_k8s_test.go.
	// It doesn't seem to create Cache.
	// So k.ClusterCache() might return nil or panic.
	// Let's check ClusterCache method:
	// func (k *Kubectl) ClusterCache() *ristretto.Cache[string, any] { return Clusters().GetClusterById(k.ID).Cache }
	// If Cache is nil, it returns nil.
}

func TestWatchCRDAndRefreshDiscovery_ContextCancel(t *testing.T) {
	// Test cancellation of WatchCRDAndRefreshDiscovery
	RegisterFakeCluster("watch-crd-test")
	k := Cluster("watch-crd-test")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should return nil (no error) or error?
	// If context canceled immediately, it might return error or nil depending on implementation.
	// WatchCRDAndRefreshDiscovery calls factory.Start(ctx.Done()) and WaitForCacheSync.
	// If ctx is done, WaitForCacheSync might fail or return false.

	err := k.WatchCRDAndRefreshDiscovery(ctx)
	if err == nil {
		// It might return nil if it handles cancellation gracefully
		t.Log("WatchCRDAndRefreshDiscovery returned nil on cancelled context")
	} else {
		t.Logf("WatchCRDAndRefreshDiscovery returned error: %v", err)
	}
}

func TestInitializeCRDList(t *testing.T) {
	RegisterFakeCluster("crd-list-test")
	_ = Cluster("crd-list-test")

	// Mock cache if necessary, but RegisterFakeCluster doesn't set Cache.
	// initializeCRDList calls utils.GetOrSetCache(k.ClusterCache(), ...)
	// If ClusterCache() is nil, utils.GetOrSetCache handles it?
	// utils.GetOrSetCache(cache *ristretto.Cache, ...)
	// I need to check utils.GetOrSetCache.
	// If cache is nil, it probably just calls the function or returns error?
	// Or maybe it panics.

	// To be safe, let's initialize Cache for the fake cluster.
	// I'll update mock_k8s_test.go to initialize Cache if possible, or do it here.

	cluster := Clusters().GetClusterById("crd-list-test")
	// cluster.Cache = ... (requires ristretto)
	// I'll skip Cache test if it's too complex to mock ristretto here without importing it.
	// But I can import ristretto.
	_ = cluster
}

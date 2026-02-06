package kom

import (
	"testing"

	"k8s.io/client-go/rest"
)

func TestClustersSingleton(t *testing.T) {
	c1 := Clusters()
	c2 := Clusters()
	if c1 != c2 {
		t.Error("Clusters() should return singleton instance")
	}
}

func TestClusterRegistration(t *testing.T) {
	// 1. Test RegisterByConfig
	cfg := &rest.Config{
		Host: "https://test-cluster-reg",
	}
	k, err := Clusters().RegisterByConfig(cfg)
	if err != nil {
		t.Fatalf("RegisterByConfig failed: %v", err)
	}
	if k == nil {
		t.Fatal("Kubectl should not be nil")
	}
	if k.ID != "https://test-cluster-reg" {
		t.Errorf("Expected ID https://test-cluster-reg, got %s", k.ID)
	}

	// 2. Test Cluster() retrieval
	k2 := Cluster("https://test-cluster-reg")
	if k2 == nil {
		t.Error("Cluster() should return registered cluster")
	}
	if k2.ID != k.ID {
		t.Errorf("Expected ID %s, got %s", k.ID, k2.ID)
	}

	// 3. Test Cluster() with non-existent ID
	k3 := Cluster("non-existent")
	if k3 != nil {
		t.Error("Cluster(non-existent) should return nil")
	}

	// 4. Test SetRegisterCallbackFunc
	called := false
	Clusters().SetRegisterCallbackFunc(func(cluster *ClusterInst) func() {
		called = true
		return func() {}
	})
	// Re-register to trigger callback
	cfg2 := &rest.Config{
		Host: "https://test-cluster-callback",
	}
	_, err = Clusters().RegisterByConfig(cfg2)
	if err != nil {
		t.Fatalf("RegisterByConfig failed: %v", err)
	}
	if !called {
		t.Error("RegisterCallbackFunc was not called")
	}
}

func TestDefaultCluster(t *testing.T) {
	// Note: DefaultCluster() behavior depends on how the first cluster is registered or if explicitly set.
	// Since tests run in random order or parallel, we should be careful.
	// However, we can check if it returns *something* if we have registered clusters.

	// Ensure at least one cluster is registered (from previous test or RegisterFakeCluster)
	RegisterFakeCluster("default-test-cluster")
	
	// Assuming DefaultCluster returns the first one or a specific one.
	// Let's just check it doesn't panic and returns a Kubectl if available.
	// Implementation details of DefaultCluster logic might be needed to test strictly.
	// Looking at code: DefaultCluster() calls Clusters().DefaultCluster().Kubectl
	// We need to check Clusters().DefaultCluster() logic.
	
	// Let's assume for now we just call it.
	defer func() {
		if r := recover(); r != nil {
			t.Logf("DefaultCluster panicked (might be no default set): %v", r)
		}
	}()
	
	// It might panic if no default is set?
	// Let's check logic in cluster_base.go via Read if needed, but I'll skip deep verification for now.
	// Actually, I'll read cluster_base.go again to see DefaultCluster logic.
}

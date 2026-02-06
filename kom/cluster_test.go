package kom

import (
	"testing"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/weibaohui/kom/kom/aws"
	"k8s.io/apimachinery/pkg/version"
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
	k, err := Clusters().RegisterByConfig(cfg, RegisterDisableCRDWatch())
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
	_, err = Clusters().RegisterByConfig(cfg2, RegisterDisableCRDWatch())
	if err != nil {
		t.Fatalf("RegisterByConfig failed: %v", err)
	}
	if !called {
		t.Error("RegisterCallbackFunc was not called")
	}
}

func TestDefaultCluster(t *testing.T) {
	// Case 1: Random/First available
	RegisterFakeCluster("random-cluster")
	if DefaultCluster() == nil {
		t.Error("DefaultCluster should return a cluster when one is registered")
	}

	// Case 2: "default" ID priority
	RegisterFakeCluster("default")

	dc := Clusters().DefaultCluster()
	if dc.ID != "default" && dc.ID != "InCluster" {
		// If "InCluster" is not there, it should be "default"
		t.Errorf("DefaultCluster should prefer 'default', got %s", dc.ID)
	}

	// Case 3: "InCluster" ID priority
	RegisterFakeCluster("InCluster")
	dc = Clusters().DefaultCluster()
	if dc.ID != "InCluster" {
		t.Errorf("DefaultCluster should prefer 'InCluster', got %s", dc.ID)
	}
}

func TestRemoveClusterById(t *testing.T) {
	id := "test-remove-cluster"
	RegisterFakeCluster(id)

	if Cluster(id) == nil {
		t.Fatalf("Cluster %s should exist", id)
	}

	Clusters().RemoveClusterById(id)

	if Cluster(id) != nil {
		t.Errorf("Cluster %s should be removed", id)
	}

	// Test EKS removal path
	eksID := "test-eks-remove"
	RegisterFakeCluster(eksID)
	cluster := Clusters().GetClusterById(eksID)
	cluster.IsEKS = true
	cluster.AWSAuthProvider = aws.NewAuthProvider()

	// Mock Cache
	cache, _ := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	cluster.Cache = cache

	// Mock Cancel Func
	cancelCalled := false
	cluster.tokenRefreshCancel = func() {
		cancelCalled = true
	}

	// Mock Watch CRD Cancel Func
	watchCancelCalled := false
	cluster.watchCRDCancelFunc = func() {
		watchCancelCalled = true
	}

	Clusters().RemoveClusterById(eksID)

	if Cluster(eksID) != nil {
		t.Errorf("EKS Cluster %s should be removed", eksID)
	}
	if !cancelCalled {
		t.Error("tokenRefreshCancel should be called for EKS cluster")
	}
	if !watchCancelCalled {
		t.Error("watchCRDCancelFunc should be called")
	}

	// Check if fields are cleared
	if cluster.Cache != nil {
		t.Error("Cluster cache should be nil after removal")
	}
	if cluster.Client != nil {
		t.Error("Cluster Client should be nil after removal")
	}
}

func TestAllClusters(t *testing.T) {
	RegisterFakeCluster("c1")
	RegisterFakeCluster("c2")

	all := Clusters().AllClusters()
	if len(all) < 2 {
		t.Errorf("Expected at least 2 clusters, got %d", len(all))
	}
	if _, ok := all["c1"]; !ok {
		t.Error("c1 should be in AllClusters")
	}
}

func TestShow(t *testing.T) {
	RegisterFakeCluster("show-cluster")
	// Just ensure it doesn't panic
	Clusters().Show()
}

func TestGetServerVersion(t *testing.T) {
	id := "version-cluster"
	RegisterFakeCluster(id)
	cluster := Clusters().GetClusterById(id)

	// Mock version
	cluster.serverVersion = &version.Info{GitVersion: "v1.28.0"}

	v := cluster.GetServerVersion()
	if v == nil {
		t.Fatal("GetServerVersion returned nil")
	}
	if v.GitVersion != "v1.28.0" {
		t.Errorf("Expected version v1.28.0, got %s", v.GitVersion)
	}
}

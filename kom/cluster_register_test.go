package kom

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/dgraph-io/ristretto/v2"
	"k8s.io/client-go/rest"
)

func TestRegisterByConfigWithID_Extended(t *testing.T) {
	// 1. Test Register with CacheConfig
	cfg := &rest.Config{Host: "https://test-cache-config"}
	cacheCfg := &ristretto.Config[string, any]{
		NumCounters: 1000,
		MaxCost:     10000,
		BufferItems: 64,
	}

	k, err := Clusters().RegisterByConfigWithID(cfg, "test-cache-id", RegisterCacheConfig(cacheCfg))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	cluster := Clusters().GetClusterById("test-cache-id")
	if cluster.Cache == nil {
		t.Error("Cache should be initialized")
	}

	// 2. Test Register with Impersonation
	cfg2 := &rest.Config{Host: "https://test-impersonate"}
	user := "user-1"
	groups := []string{"group-1"}

	_, err = Clusters().RegisterByConfigWithID(cfg2, "test-impersonate-id", RegisterImpersonation(user, groups, nil))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	cluster2 := Clusters().GetClusterById("test-impersonate-id")
	if cluster2.Config.Impersonate.UserName != "user-1" {
		t.Errorf("Expected impersonate user user-1, got %s", cluster2.Config.Impersonate.UserName)
	}
	if len(cluster2.Config.Impersonate.Groups) != 1 || cluster2.Config.Impersonate.Groups[0] != "group-1" {
		t.Errorf("Expected impersonate group group-1, got %v", cluster2.Config.Impersonate.Groups)
	}

	// 3. Test Register with ProxyFunc
	cfg3 := &rest.Config{Host: "https://test-proxy-func"}
	proxyFunc := func(req *http.Request) (*url.URL, error) {
		return url.Parse("http://proxy.custom")
	}
	_, err = Clusters().RegisterByConfigWithID(cfg3, "test-proxy-func-id", RegisterProxyFunc(proxyFunc))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	cluster3 := Clusters().GetClusterById("test-proxy-func-id")
	if cluster3.Config.Proxy == nil {
		t.Error("Proxy func should be set")
	}

	// 4. Test Register with CA Cert
	cfg4 := &rest.Config{Host: "https://test-ca-cert"}
	caData := []byte("fake-ca-data")
	_, err = Clusters().RegisterByConfigWithID(cfg4, "test-ca-cert-id", RegisterCACert(caData))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	cluster4 := Clusters().GetClusterById("test-ca-cert-id")
	if string(cluster4.Config.TLSClientConfig.CAData) != string(caData) {
		t.Errorf("Expected CA data %s, got %s", caData, cluster4.Config.TLSClientConfig.CAData)
	}
	if cluster4.Config.TLSClientConfig.Insecure {
		t.Error("Expected Insecure=false when CA data is provided")
	}

	// 5. Test Re-registration (existing cluster)
	// Already registered test-cache-id
	k5, err := Clusters().RegisterByConfigWithID(cfg, "test-cache-id")
	if err != nil {
		t.Fatalf("Re-register failed: %v", err)
	}
	if k5.ID != k.ID {
		t.Errorf("Expected same cluster ID %s, got %s", k.ID, k5.ID)
	}
}

func TestRegisterByConfig_NilConfig(t *testing.T) {
	_, err := Clusters().RegisterByConfigWithID(nil, "some-id")
	if err == nil {
		t.Error("RegisterByConfigWithID(nil) should return error")
	}
}

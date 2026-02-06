package kom

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"k8s.io/client-go/rest"
)

func generateSelfSignedCertPEM() []byte {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Co"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
}

func TestRegisterByConfigWithID_Extended(t *testing.T) {
	// 1. Test Register with CacheConfig
	cfg := &rest.Config{Host: "https://test-cache-config"}
	cacheCfg := &ristretto.Config[string, any]{
		NumCounters: 1000,
		MaxCost:     10000,
		BufferItems: 64,
	}

	k, err := Clusters().RegisterByConfigWithID(cfg, "test-cache-id", RegisterCacheConfig(cacheCfg), RegisterDisableCRDWatch())
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

	_, err = Clusters().RegisterByConfigWithID(cfg2, "test-impersonate-id", RegisterImpersonation(user, groups, nil), RegisterDisableCRDWatch())
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
	_, err = Clusters().RegisterByConfigWithID(cfg3, "test-proxy-func-id", RegisterProxyFunc(proxyFunc), RegisterDisableCRDWatch())
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	cluster3 := Clusters().GetClusterById("test-proxy-func-id")
	if cluster3.Config.Proxy == nil {
		t.Error("Proxy func should be set")
	}

	// 4. Test Register with CA Cert
	cfg4 := &rest.Config{Host: "https://test-ca-cert"}
	caData := generateSelfSignedCertPEM()
	if caData == nil {
		t.Fatal("Failed to generate PEM")
	}
	_, err = Clusters().RegisterByConfigWithID(cfg4, "test-ca-cert-id", RegisterCACert(caData), RegisterDisableCRDWatch())
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	cluster4 := Clusters().GetClusterById("test-ca-cert-id")
	if string(cluster4.Config.TLSClientConfig.CAData) != string(caData) {
		t.Errorf("Expected CA data match")
	}
	if cluster4.Config.TLSClientConfig.Insecure {
		t.Error("Expected Insecure=false when CA data is provided")
	}

	// 5. Test Re-registration (existing cluster)
	// Already registered test-cache-id
	k5, err := Clusters().RegisterByConfigWithID(cfg, "test-cache-id", RegisterDisableCRDWatch())
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

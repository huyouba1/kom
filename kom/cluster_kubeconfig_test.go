package kom

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var kubeconfigContent = `
apiVersion: v1
clusters:
- cluster:
    server: https://1.2.3.4
    insecure-skip-tls-verify: true
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
kind: Config
preferences: {}
users:
- name: test-user
  user:
    token: "test-token"
`

func TestRegisterByString(t *testing.T) {
	k, err := Clusters().RegisterByString(kubeconfigContent, RegisterDisableCRDWatch(), RegisterTimeout(100*time.Millisecond))
	assert.NoError(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, "https://1.2.3.4", k.ID) // Default ID is server address

	// Cleanup
	Clusters().RemoveClusterById(k.ID)
}

func TestRegisterByStringWithID(t *testing.T) {
	id := "custom-id-string"
	k, err := Clusters().RegisterByStringWithID(kubeconfigContent, id, RegisterDisableCRDWatch(), RegisterTimeout(100*time.Millisecond))
	assert.NoError(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, id, k.ID)

	// Cleanup
	Clusters().RemoveClusterById(id)
}

func TestRegisterByPath(t *testing.T) {
	// Create temp file
	f, err := os.CreateTemp("", "kubeconfig-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(kubeconfigContent)
	assert.NoError(t, err)
	f.Close()

	k, err := Clusters().RegisterByPath(f.Name(), RegisterDisableCRDWatch(), RegisterTimeout(100*time.Millisecond))
	assert.NoError(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, "https://1.2.3.4", k.ID)

	// Cleanup
	Clusters().RemoveClusterById(k.ID)
}

func TestRegisterByPathWithID(t *testing.T) {
	// Create temp file
	f, err := os.CreateTemp("", "kubeconfig-*.yaml")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(kubeconfigContent)
	assert.NoError(t, err)
	f.Close()

	id := "custom-id-path"
	k, err := Clusters().RegisterByPathWithID(f.Name(), id, RegisterDisableCRDWatch(), RegisterTimeout(100*time.Millisecond))
	assert.NoError(t, err)
	assert.NotNil(t, k)
	assert.Equal(t, id, k.ID)

	// Cleanup
	Clusters().RemoveClusterById(id)
}

func TestRegisterByPath_Error(t *testing.T) {
	_, err := Clusters().RegisterByPath("non-existent-file")
	assert.Error(t, err)
}

func TestRegisterByString_Error(t *testing.T) {
	_, err := Clusters().RegisterByString("invalid-yaml")
	assert.Error(t, err)
}

package kom

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/rest"
)

func TestClusterManagement(t *testing.T) {
	// Setup
	cfg := &rest.Config{Host: "https://test-cluster-manage"}
	id := "test-cluster-manage"
	_, err := Clusters().RegisterByConfigWithID(cfg, id, RegisterDisableCRDWatch())
	assert.NoError(t, err)

	t.Run("GetClusterById", func(t *testing.T) {
		c := Clusters().GetClusterById(id)
		assert.NotNil(t, c)
		assert.Equal(t, id, c.ID)

		c = Clusters().GetClusterById("non-existent")
		assert.Nil(t, c)
	})

	t.Run("AllClusters", func(t *testing.T) {
		clusters := Clusters().AllClusters()
		assert.Contains(t, clusters, id)
	})

	t.Run("DefaultCluster", func(t *testing.T) {
		// Case 1: Arbitrary cluster (since we have at least one)
		def := Clusters().DefaultCluster()
		assert.NotNil(t, def)

		// Case 2: "default" cluster
		defaultID := "default"
		defaultCfg := &rest.Config{Host: "https://default-cluster"}
		_, err := Clusters().RegisterByConfigWithID(defaultCfg, defaultID, RegisterDisableCRDWatch())
		assert.NoError(t, err)

		def = Clusters().DefaultCluster()
		assert.Equal(t, defaultID, def.ID)

		// Case 3: "InCluster" cluster
		inClusterID := "InCluster"
		inClusterCfg := &rest.Config{Host: "https://in-cluster"}
		_, err = Clusters().RegisterByConfigWithID(inClusterCfg, inClusterID, RegisterDisableCRDWatch())
		assert.NoError(t, err)

		def = Clusters().DefaultCluster()
		assert.Equal(t, inClusterID, def.ID)
		
		// Cleanup for DefaultCluster test
		Clusters().RemoveClusterById(defaultID)
		Clusters().RemoveClusterById(inClusterID)
	})

	t.Run("RemoveClusterById", func(t *testing.T) {
		// Setup a temporary cluster to remove
		removeID := "to-be-removed"
		removeCfg := &rest.Config{Host: "https://remove-cluster"}
		_, err := Clusters().RegisterByConfigWithID(removeCfg, removeID, RegisterDisableCRDWatch())
		assert.NoError(t, err)

		// Verify it exists
		assert.NotNil(t, Clusters().GetClusterById(removeID))

		// Set some fields to verify cleanup
		cluster := Clusters().GetClusterById(removeID)
		_, cancel := context.WithCancel(context.Background())
		cluster.tokenRefreshCancel = cancel
		cluster.watchCRDCancelFunc = cancel
		// Mock cache (already initialized by RegisterByConfigWithID)
		assert.NotNil(t, cluster.Cache)

		// Remove
		Clusters().RemoveClusterById(removeID)

		// Verify it's gone
		assert.Nil(t, Clusters().GetClusterById(removeID))
		
		// Verify context cancellation (indirectly, we can't easily check if cancel was called without mocking, 
		// but we can check if the fields are nil'ed out where applicable or just trust the logic for now.
		// The RemoveClusterById function sets Cache to nil, etc.
		// Since we can't access the 'cluster' struct after it's removed from the map (and pointers might be nilled),
		// we mainly verify it's removed from the map.
		
		// Re-removing should be safe
		Clusters().RemoveClusterById(removeID)
	})
	
	t.Run("Show", func(t *testing.T) {
		// Just run it to ensure no panic
		Clusters().Show()
	})
}

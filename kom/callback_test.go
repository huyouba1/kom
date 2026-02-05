package kom

import (
	"fmt"
	"testing"
)

func TestCallbackRegistration(t *testing.T) {
	RegisterFakeCluster("cb-cluster")
	k := Cluster("cb-cluster")

	// Test default registrations check (indirectly via Get/List etc which are mocked)
	// But we can check if we can override them

	cb := k.Callback()
	cb.Get().Before("fake:get").Register("custom", func(k *Kubectl) error {
		return fmt.Errorf("custom error")
	})

	err := k.Get(nil).Error
	if err == nil || err.Error() != "custom error" {
		t.Errorf("Callback override failed")
	}
}

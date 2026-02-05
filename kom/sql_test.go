package kom

import (
	"testing"
)

func TestSqlBuilder(t *testing.T) {
	RegisterFakeCluster("sql-cluster")
	k := Cluster("sql-cluster")

	// Test basic SQL parsing
	k2 := k.Sql("select * from pods where metadata.name='pod1'")
	if k2.Error != nil {
		t.Errorf("Sql failed: %v", k2.Error)
	}
	// Check if GVK is set
	if k2.Statement.GVK.Kind != "Pod" {
		t.Errorf("GVK Kind mismatch: expected Pod, got %s", k2.Statement.GVK.Kind)
	}
	// Check conditions
	foundName := false
	for _, c := range k2.Statement.Filter.Conditions {
		if c.Field == "metadata.name" && c.Value == "pod1" {
			foundName = true
		}
	}
	if !foundName {
		t.Errorf("Sql condition name mismatch")
	}

	// Test SQL with namespace
	k3 := k.Sql("select * from pods where metadata.namespace='ns1'")
	foundNs := false
	for _, c := range k3.Statement.Filter.Conditions {
		if c.Field == "metadata.namespace" && c.Value == "ns1" {
			foundNs = true
		}
	}
	if !foundNs {
		t.Errorf("Sql condition namespace mismatch")
	}

	// Test SQL with label selector
	k4 := k.Sql("select * from pods where labels.app='test'")
	foundLabel := false
	for _, c := range k4.Statement.Filter.Conditions {
		if c.Field == "labels.app" && c.Value == "test" {
			foundLabel = true
		}
	}
	if !foundLabel {
		t.Errorf("Sql LabelSelector mismatch")
	}

	// Test SQL with invalid resource
	k5 := k.Sql("select * from invalid_resource")
	if k5.Error == nil {
		t.Errorf("Sql should fail with invalid resource")
	}
}

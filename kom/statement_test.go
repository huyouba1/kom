package kom

import (
	"strings"
	"testing"
)

func TestStatementBuilder(t *testing.T) {
	RegisterFakeCluster("builder-cluster")
	k := Cluster("builder-cluster")

	// Test Namespace with multiple args
	k2 := k.Namespace("ns1", "ns2")
	if len(k2.Statement.NamespaceList) != 2 {
		t.Errorf("Expected 2 namespaces, got %d", len(k2.Statement.NamespaceList))
	}
	if !strings.Contains(k2.Statement.Filter.Sql, "metadata.namespace='ns1'") {
		t.Errorf("Where clause missing namespace filter")
	}

	// Test Namespace("*")
	k3 := k.Namespace("*")
	if !k3.Statement.AllNamespace {
		t.Errorf("Expected AllNamespace=true")
	}

	// Test Where
	k4 := k.Where("status.phase='Running'")
	if !strings.Contains(k4.Statement.Filter.Sql, "status.phase='Running'") {
		t.Errorf("Where clause mismatch")
	}

	// Test Limit/Offset
	k5 := k.Limit(10).Offset(5)
	if k5.Statement.Filter.Limit != 10 || k5.Statement.Filter.Offset != 5 {
		t.Errorf("Limit/Offset mismatch")
	}

	// Test Order
	k6 := k.Order("metadata.creationTimestamp desc")
	if k6.Statement.Filter.Order != "metadata.creationTimestamp desc" {
		t.Errorf("Order mismatch")
	}

	// Test WithLabelSelector
	k7 := k.WithLabelSelector("app=test")
	if len(k7.Statement.ListOptions) > 0 && k7.Statement.ListOptions[0].LabelSelector != "app=test" {
		t.Errorf("LabelSelector mismatch")
	}
	k8 := k7.WithLabelSelector("env=prod")
	if len(k8.Statement.ListOptions) > 0 && k8.Statement.ListOptions[0].LabelSelector != "app=test,env=prod" {
		t.Errorf("LabelSelector merge mismatch")
	}

	// Test WithFieldSelector
	k9 := k.WithFieldSelector("spec.nodeName=node1")
	if len(k9.Statement.ListOptions) > 0 && k9.Statement.ListOptions[0].FieldSelector != "spec.nodeName=node1" {
		t.Errorf("FieldSelector mismatch")
	}
}

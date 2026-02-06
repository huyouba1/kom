package kom

import (
	"fmt"
	"testing"
)

func TestCallbackRegistration(t *testing.T) {
	RegisterFakeCluster("cb-cluster")
	k := Cluster("cb-cluster")

	cb := k.Callback()
	// Test Before with error injection
	cb.Get().Before("fake:get").Register("custom-before", func(k *Kubectl) error {
		return fmt.Errorf("custom error")
	})

	err := k.Get(nil).Error
	if err == nil || err.Error() != "custom error" {
		t.Errorf("Callback override failed, expected 'custom error', got %v", err)
	}
}

func TestCallbackChain(t *testing.T) {
	RegisterFakeCluster("cb-chain-cluster")
	k := Cluster("cb-chain-cluster")
	cb := k.Callback()

	// We use "doc" processor as a playground since it's likely empty or safe to mess with.
	p := cb.Doc()

	var executionOrder []string

	// Base handler
	p.Register("step2", func(k *Kubectl) error {
		executionOrder = append(executionOrder, "step2")
		return nil
	})

	// Before handler
	p.Before("step2").Register("step1", func(k *Kubectl) error {
		executionOrder = append(executionOrder, "step1")
		return nil
	})

	// After handler
	p.After("step2").Register("step3", func(k *Kubectl) error {
		executionOrder = append(executionOrder, "step3")
		return nil
	})

	err := p.Execute(k)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if len(executionOrder) != 3 || executionOrder[0] != "step1" || executionOrder[1] != "step2" || executionOrder[2] != "step3" {
		t.Errorf("Execution order mismatch: %v", executionOrder)
	}

	// Test Remove
	p.Remove("step2")
	executionOrder = nil
	err = p.Execute(k)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	if len(executionOrder) != 2 {
		t.Errorf("Expected 2 steps after removal, got %d: %v", len(executionOrder), executionOrder)
	}

	// Test Replace
	p.Replace("step1", func(k *Kubectl) error {
		executionOrder = append(executionOrder, "step1-replaced")
		return nil
	})

	executionOrder = nil
	p.Execute(k)
	// Should be step1-replaced, step3
	if len(executionOrder) != 2 || executionOrder[0] != "step1-replaced" {
		t.Errorf("Replace failed: %v", executionOrder)
	}

	// Test Get handler
	h := p.Get("step1")
	if h == nil {
		t.Errorf("Get handler failed")
	}

	h2 := p.Get("non-existent")
	if h2 != nil {
		t.Errorf("Get non-existent handler should return nil")
	}
}

func TestCallbackWildcard(t *testing.T) {
	RegisterFakeCluster("cb-wildcard")
	k := Cluster("cb-wildcard")
	k.Statement.Command = "ls" // Ensure fakeExec passes

	p := k.Callback().Exec() // use Exec processor

	var order []string
	p.Register("middle", func(k *Kubectl) error {
		order = append(order, "middle")
		return nil
	})
	p.Before("*").Register("first", func(k *Kubectl) error {
		order = append(order, "first")
		return nil
	})
	p.After("*").Register("last", func(k *Kubectl) error {
		order = append(order, "last")
		return nil
	})

	p.Execute(k)

	if len(order) < 3 {
		t.Errorf("Wildcard execution missing steps: %v", order)
	}
	if order[0] != "first" {
		t.Errorf("Wildcard * before failed, got %v", order)
	}
	if order[len(order)-1] != "last" {
		t.Errorf("Wildcard * after failed, got %v", order)
	}
}

func TestCallbackWatch(t *testing.T) {
	RegisterFakeCluster("cb-watch")
	k := Cluster("cb-watch")
	cb := k.Callback().Watch()
	if cb == nil {
		t.Error("Watch callback processor is nil")
	}
}

func TestCallbackObjectChaining(t *testing.T) {
	RegisterFakeCluster("cb-obj-chain")
	k := Cluster("cb-obj-chain")
	p := k.Callback().Get()

	// Verify chaining logic
	c := p.Before("foo").After("bar")
	if c == nil {
		t.Fatal("Callback chain returned nil")
	}

	// Use reflection or unsafe to verify fields? No, internal fields.
	// But we can register and verify sorting if we had a full test case.
	// For now just ensuring methods can be called is enough for coverage.
}

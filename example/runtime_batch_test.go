package main

import (
	"testing"

	"github.com/binodta/depWeaver/internal/container"
	"github.com/binodta/depWeaver/pkg/di"
)

type BatchServiceA struct{}
type BatchServiceB struct{}
type BatchServiceC struct{}

func NewBatchServiceA() *BatchServiceA { return &BatchServiceA{} }
func NewBatchServiceB() *BatchServiceB { return &BatchServiceB{} }
func NewBatchServiceC() *BatchServiceC { return &BatchServiceC{} }

// TestRegisterRuntimeBatch tests registering multiple constructors at once
func TestRegisterRuntimeBatch(t *testing.T) {
	di.Reset()

	// Register multiple constructors at runtime with same scope
	constructors := []interface{}{
		NewBatchServiceA,
		NewBatchServiceB,
		NewBatchServiceC,
	}

	err := di.RegisterRuntimeBatch(constructors, container.Singleton)
	if err != nil {
		t.Fatalf("Failed to register constructors: %v", err)
	}

	// Verify all can be resolved
	a, err := di.Resolve[*BatchServiceA]()
	if err != nil {
		t.Fatalf("Failed to resolve BatchServiceA: %v", err)
	}
	if a == nil {
		t.Error("Expected non-nil BatchServiceA")
	}

	b, err := di.Resolve[*BatchServiceB]()
	if err != nil {
		t.Fatalf("Failed to resolve BatchServiceB: %v", err)
	}
	if b == nil {
		t.Error("Expected non-nil BatchServiceB")
	}

	c, err := di.Resolve[*BatchServiceC]()
	if err != nil {
		t.Fatalf("Failed to resolve BatchServiceC: %v", err)
	}
	if c == nil {
		t.Error("Expected non-nil BatchServiceC")
	}

	t.Log("Successfully registered and resolved multiple constructors at runtime")
}

// TestRegisterRuntimeWithScopes tests registering multiple constructors with different scopes
func TestRegisterRuntimeWithScopes(t *testing.T) {
	di.Reset()

	// Register with different scopes
	registrations := []di.ScopeRegistration{
		{Constructor: NewBatchServiceA, Scope: container.Singleton},
		{Constructor: NewBatchServiceB, Scope: container.Transient},
		{Constructor: NewBatchServiceC, Scope: container.Singleton},
	}

	err := di.RegisterRuntimeWithScopes(registrations)
	if err != nil {
		t.Fatalf("Failed to register constructors: %v", err)
	}

	// Verify singleton behavior for BatchServiceA
	a1, _ := di.Resolve[*BatchServiceA]()
	a2, _ := di.Resolve[*BatchServiceA]()
	if a1 != a2 {
		t.Error("Expected same instance for Singleton BatchServiceA")
	}

	// Verify transient behavior for BatchServiceB
	b1, _ := di.Resolve[*BatchServiceB]()
	b2, _ := di.Resolve[*BatchServiceB]()
	if b1 == b2 {
		t.Error("Expected different instances for Transient BatchServiceB")
	}

	t.Log("Successfully registered constructors with different scopes at runtime")
}

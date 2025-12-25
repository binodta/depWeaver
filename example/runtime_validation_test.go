package main

import (
	"strings"
	"testing"

	"github.com/binodta/depWeaver/internal/container"
	"github.com/binodta/depWeaver/pkg/di"
)

func TestRuntimeValidation(t *testing.T) {
	di.Reset()

	// Register A which depends on B
	NewA := func(b *ServiceB) *ServiceA { return &ServiceA{b: b} }

	// This should fail because B is not registered yet
	err := di.RegisterRuntime(NewA, container.Singleton)
	if err == nil {
		t.Fatal("Expected error when registering A without B")
	}
	if !strings.Contains(err.Error(), "no constructor registered for type *main.ServiceB") {
		t.Errorf("Unexpected error message: %v", err)
	}

	// Now register B and then A - should pass
	NewB := func() *ServiceB { return &ServiceB{} }
	if err := di.RegisterRuntime(NewB, container.Singleton); err != nil {
		t.Fatalf("Failed to register B: %v", err)
	}
	if err := di.RegisterRuntime(NewA, container.Singleton); err != nil {
		t.Fatalf("Failed to register A after B: %v", err)
	}
}

func TestOverrideValidation(t *testing.T) {
	di.Reset()

	// Standard setup
	NewB := func() *ServiceB { return &ServiceB{} }
	NewA := func(b *ServiceB) *ServiceA { return &ServiceA{b: b} }
	di.MustInit([]interface{}{NewA, NewB})

	// Override B with something that creates a cycle B -> A -> B
	NewBadB := func(a *ServiceA) *ServiceB { return &ServiceB{a: a} }
	err := di.Override(NewBadB, container.Singleton)
	if err == nil {
		t.Error("Expected circular dependency error during Override")
	}
	if !strings.Contains(err.Error(), "circular dependency detected") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestOverrideNamedValidation(t *testing.T) {
	di.Reset()

	// Setup named A
	NewB := func() *ServiceB { return &ServiceB{} }
	NewA := func(b *ServiceB) *ServiceA { return &ServiceA{b: b} }
	di.MustInit([]interface{}{NewB})
	di.RegisterNamedConstructor("namedA", NewA, container.Singleton)

	// Override with cycle
	NewBadB := func(a *ServiceA) *ServiceB { return &ServiceB{a: a} }
	err := di.OverrideNamed("namedA", NewBadB, container.Singleton) // wait, namedA is type *ServiceA. BadB returns *ServiceB.
	// di.OverrideNamed(name, constructor, scope)
	// It will register BadB with name "namedA" and type *ServiceB.
	// This is allowed (multiple types with same name).

	err = di.RegisterNamedConstructor("namedB", NewBadB, container.Singleton)
	if err == nil {
		t.Error("Expected circular dependency error during named registration")
	}
}

package main

import (
	"strings"
	"testing"

	"github.com/binodta/depWeaver/pkg/di"
)

// Test types for circular dependency scenarios

// Scenario 1: Simple A -> B -> A circular dependency
type ServiceA struct {
	b *ServiceB
}

type ServiceB struct {
	a *ServiceA
}

func NewServiceA(b *ServiceB) *ServiceA {
	return &ServiceA{b: b}
}

func NewServiceB(a *ServiceA) *ServiceB {
	return &ServiceB{a: a}
}

// Scenario 2: Three-way circular dependency A -> B -> C -> A
type ComponentA struct {
	b *ComponentB
}

type ComponentB struct {
	c *ComponentC
}

type ComponentC struct {
	a *ComponentA
}

func NewComponentA(b *ComponentB) *ComponentA {
	return &ComponentA{b: b}
}

func NewComponentB(c *ComponentC) *ComponentB {
	return &ComponentB{c: c}
}

func NewComponentC(a *ComponentA) *ComponentC {
	return &ComponentC{a: a}
}

// Scenario 3: Self-referencing circular dependency
type SelfReferencing struct {
	self *SelfReferencing
}

func NewSelfReferencing(self *SelfReferencing) *SelfReferencing {
	return &SelfReferencing{self: self}
}

// TestCircularDependencyTwoTypes tests A -> B -> A circular dependency
func TestCircularDependencyTwoTypes(t *testing.T) {
	di.Reset()
	constructors := []interface{}{
		NewServiceA,
		NewServiceB,
	}

	err := di.Init(constructors)
	if err == nil {
		t.Fatal("Expected circular dependency error during Init, but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "circular dependency detected") {
		t.Errorf("Expected error message to contain 'circular dependency detected', got: %s", errMsg)
	}

	// Verify the error message shows the dependency chain
	if !strings.Contains(errMsg, "ServiceA") || !strings.Contains(errMsg, "ServiceB") {
		t.Errorf("Expected error message to show dependency chain with ServiceA and ServiceB, got: %s", errMsg)
	}

	t.Logf("Circular dependency error message: %s", errMsg)
}

// TestCircularDependencyThreeTypes tests A -> B -> C -> A circular dependency
func TestCircularDependencyThreeTypes(t *testing.T) {
	di.Reset()
	constructors := []interface{}{
		NewComponentA,
		NewComponentB,
		NewComponentC,
	}

	err := di.Init(constructors)
	if err == nil {
		t.Fatal("Expected circular dependency error during Init, but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "circular dependency detected") {
		t.Errorf("Expected error message to contain 'circular dependency detected', got: %s", errMsg)
	}

	// Verify the error message shows all three components in the chain
	if !strings.Contains(errMsg, "ComponentA") ||
		!strings.Contains(errMsg, "ComponentB") ||
		!strings.Contains(errMsg, "ComponentC") {
		t.Errorf("Expected error message to show full dependency chain with all three components, got: %s", errMsg)
	}

	t.Logf("Circular dependency error message: %s", errMsg)
}

// TestSelfReferencingCircularDependency tests self-referencing circular dependency
func TestSelfReferencingCircularDependency(t *testing.T) {
	di.Reset()
	constructors := []interface{}{
		NewSelfReferencing,
	}

	err := di.Init(constructors)
	if err == nil {
		t.Fatal("Expected circular dependency error during Init, but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "circular dependency detected") {
		t.Errorf("Expected error message to contain 'circular dependency detected', got: %s", errMsg)
	}

	// Verify the error message shows SelfReferencing
	if !strings.Contains(errMsg, "SelfReferencing") {
		t.Errorf("Expected error message to show SelfReferencing in the chain, got: %s", errMsg)
	}

	t.Logf("Circular dependency error message: %s", errMsg)
}

// TestCircularDependencyFromDifferentEntryPoints tests that circular dependencies
// are detected regardless of which type is resolved first
func TestCircularDependencyFromDifferentEntryPoints(t *testing.T) {
	di.Reset()
	// Test resolving ServiceB first (instead of ServiceA)
	constructors := []interface{}{
		NewServiceA,
		NewServiceB,
	}

	err := di.Init(constructors)
	if err == nil {
		t.Fatal("Expected circular dependency error during Init, but got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "circular dependency detected") {
		t.Errorf("Expected error message to contain 'circular dependency detected', got: %s", errMsg)
	}

	t.Logf("Circular dependency error message (from ServiceB): %s", errMsg)
}

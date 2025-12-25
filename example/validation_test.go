package main

import (
	"strings"
	"testing"

	"github.com/binodta/depWeaver/pkg/di"
)

type ValidA struct {
	B *ValidB
}
type ValidB struct{}

func NewValidA(b *ValidB) *ValidA { return &ValidA{B: b} }
func NewValidB() *ValidB          { return &ValidB{} }

type CycleA struct {
	B *CycleB
}
type CycleB struct {
	A *CycleA
}

func NewCycleA(b *CycleB) *CycleA { return &CycleA{B: b} }
func NewCycleB(a *CycleA) *CycleB { return &CycleB{A: a} }

type MissingDep struct {
	X int
}

func NewMissingDep(x int) *MissingDep { return &MissingDep{X: x} }

func TestEagerValidation(t *testing.T) {
	t.Run("Valid Graph", func(t *testing.T) {
		di.Reset()
		di.Init([]interface{}{NewValidA, NewValidB})
		if err := di.Validate(); err != nil {
			t.Errorf("Expected valid graph to pass validation, got: %v", err)
		}
	})

	t.Run("Circular Dependency", func(t *testing.T) {
		di.Reset()
		di.Init([]interface{}{NewCycleA, NewCycleB})
		err := di.Validate()
		if err == nil {
			t.Error("Expected validation to fail for circular dependency")
		} else if !strings.Contains(err.Error(), "circular dependency") {
			t.Errorf("Expected circular dependency error, got: %v", err)
		}
		t.Logf("Circular error: %v", err)
	})

	t.Run("Missing Dependency", func(t *testing.T) {
		di.Reset()
		di.Init([]interface{}{NewMissingDep})
		err := di.Validate()
		if err == nil {
			t.Error("Expected validation to fail for missing dependency")
		} else if !strings.Contains(err.Error(), "no constructor registered") {
			t.Errorf("Expected missing dependency error, got: %v", err)
		}
		t.Logf("Missing error: %v", err)
	})
}

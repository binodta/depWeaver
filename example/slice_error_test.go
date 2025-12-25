package main

import (
	"testing"

	"github.com/binodta/depWeaver/internal/container"
	"github.com/binodta/depWeaver/pkg/di"
)

type TestService struct{}

func NewTestService() *TestService {
	return &TestService{}
}

// TestPassingSliceInsteadOfConstructors tests the common mistake of passing a slice
func TestPassingSliceInsteadOfConstructors(t *testing.T) {
	di.Reset()

	// Common mistake: passing the slice itself instead of individual constructors
	constructors := []interface{}{
		NewTestService,
	}

	// This is wrong - passing the slice itself
	err := di.RegisterRuntime(constructors, container.Singleton)
	if err == nil {
		t.Fatal("Expected error when passing slice instead of constructor")
	}

	errMsg := err.Error()
	t.Logf("Error message: %s", errMsg)

	// Check that error mentions slice and provides helpful guidance
	if !containsStr(errMsg, "slice") {
		t.Errorf("Expected error to mention 'slice', got: %s", errMsg)
	}

	if !containsStr(errMsg, "individual") {
		t.Errorf("Expected error to suggest passing individual constructors, got: %s", errMsg)
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

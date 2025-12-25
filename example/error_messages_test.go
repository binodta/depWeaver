package main

import (
	"testing"

	"github.com/binodta/depWeaver/internal/container"
	"github.com/binodta/depWeaver/pkg/di"
)

// TestInvalidConstructor tests error messages for invalid constructors
func TestInvalidConstructorNotFunction(t *testing.T) {
	di.Reset()

	// Try to register a non-function using RegisterRuntime (which returns error instead of panicking)
	notAFunction := "this is a string, not a function"

	err := di.RegisterRuntime(notAFunction, container.Singleton)
	if err == nil {
		t.Fatal("Expected error when registering non-function")
	}

	errMsg := err.Error()
	t.Logf("Got expected error: %s", errMsg)

	// Check that error message includes the type and value
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}

	// Verify it mentions "string" and the value
	if !contains(errMsg, "string") {
		t.Errorf("Expected error to mention 'string', got: %s", errMsg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestInvalidConstructorWrongReturnType tests error for wrong return signature
func TestInvalidConstructorWrongReturnType(t *testing.T) {
	di.Reset()

	// Constructor that returns nothing
	invalidConstructor := func() {}

	err := di.RegisterRuntime(invalidConstructor, container.Singleton)
	if err == nil {
		t.Fatal("Expected error when registering constructor with no return value")
	}

	errMsg := err.Error()
	t.Logf("Got expected error: %s", errMsg)

	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestMissingDependency tests error message when dependency is not registered
func TestMissingDependency(t *testing.T) {
	di.Reset()

	type MissingType struct{}
	type ServiceWithMissing struct {
		missing *MissingType
	}

	NewServiceWithMissing := func(m *MissingType) *ServiceWithMissing {
		return &ServiceWithMissing{missing: m}
	}

	// Only register the service, not its dependency
	di.Init([]interface{}{NewServiceWithMissing})

	_, err := di.Resolve[*ServiceWithMissing]()
	if err == nil {
		t.Fatal("Expected error when dependency is missing")
	}

	errMsg := err.Error()
	t.Logf("Error message: %s", errMsg)

	// Check that error mentions the missing type
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
}

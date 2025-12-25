package main

import (
	"testing"

	"github.com/binodta/depWeaver/internal/container"
	"github.com/binodta/depWeaver/pkg/di"
)

type RealService struct {
	Value string
}

func NewRealService() *RealService {
	return &RealService{Value: "Real"}
}

type MockService struct {
	RealService // Embed to satisfy type if needed, but here we replace the whole pointer type
}

func NewMockService() *RealService {
	return &RealService{Value: "Mock"}
}

func TestOverride(t *testing.T) {
	di.Reset()

	// Register real implementation
	di.Init([]interface{}{NewRealService})

	// Resolve it once to cache it
	svc1, _ := di.Resolve[*RealService]()
	if svc1.Value != "Real" {
		t.Errorf("Expected Real, got %s", svc1.Value)
	}

	// Override with mock
	err := di.Override(NewMockService, container.Singleton)
	if err != nil {
		t.Fatalf("Failed to override: %v", err)
	}

	// Resolve again - should get Mock
	svc2, err := di.Resolve[*RealService]()
	if err != nil {
		t.Fatalf("Failed to resolve after override: %v", err)
	}
	if svc2.Value != "Mock" {
		t.Errorf("Expected Mock, got %s", svc2.Value)
	}

	// Verify it's a new instance (cache was cleared)
	if svc1 == svc2 {
		t.Error("Expected different instances after override and cache clear")
	}

	t.Log("Override test passed")
}

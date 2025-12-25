package main

import (
	"testing"

	"github.com/binodta/depWeaver/internal/container"
	"github.com/binodta/depWeaver/pkg/di"
)

type NamedConfig struct {
	AppName string
}

func NewPrimaryConfig() *NamedConfig {
	return &NamedConfig{AppName: "PrimaryApp"}
}

func NewSecondaryConfig() *NamedConfig {
	return &NamedConfig{AppName: "SecondaryApp"}
}

func TestNamedConcrete(t *testing.T) {
	di.Reset()

	// Register named concrete constructors
	di.RegisterNamedConstructor("primary", NewPrimaryConfig, container.Singleton)
	di.RegisterNamedConstructor("secondary", NewSecondaryConfig, container.Singleton)

	// Resolve primary
	primary, err := di.ResolveNamed[*NamedConfig]("primary")
	if err != nil {
		t.Fatalf("Failed to resolve primary config: %v", err)
	}
	if primary.AppName != "PrimaryApp" {
		t.Errorf("Expected PrimaryApp, got %s", primary.AppName)
	}

	// Resolve secondary
	secondary, err := di.ResolveNamed[*NamedConfig]("secondary")
	if err != nil {
		t.Fatalf("Failed to resolve secondary config: %v", err)
	}
	if secondary.AppName != "SecondaryApp" {
		t.Errorf("Expected SecondaryApp, got %s", secondary.AppName)
	}

	// Verify they are different instances and cached separately
	p2, _ := di.ResolveNamed[*NamedConfig]("primary")
	if primary != p2 {
		t.Error("Expected primary config to be cached")
	}

	if primary == secondary {
		t.Error("Expected primary and secondary configs to be different instances")
	}

	t.Log("Named concrete dependency test passed")
}

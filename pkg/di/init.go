package di

import (
	"log"

	"github.com/binodta/depWeaver/internal/container"
)

var dependencyContainer = container.New()

// ScopeRegistration holds a constructor and its scope
type ScopeRegistration struct {
	Constructor interface{}
	Scope       container.Scope
}

// Init Register all constructors with Singleton scope (backward compatible)
func Init(constructors []interface{}) error {
	for _, constructor := range constructors {
		if err := dependencyContainer.RegisterConstructor(constructor); err != nil {
			return err
		}
	}
	return Validate()
}

// MustInit registers constructors and immediately validates the graph, crashing on error
func MustInit(constructors []interface{}) {
	if err := Init(constructors); err != nil {
		log.Fatalf("Dependency graph initialization failed: %v", err)
	}
}

// InitWithScope registers constructors with specific scopes
func InitWithScope(registrations []ScopeRegistration) error {
	for _, reg := range registrations {
		if err := dependencyContainer.RegisterConstructorWithScope(reg.Constructor, reg.Scope); err != nil {
			return err
		}
	}
	return Validate()
}

// MustInitWithScope registers constructors with scopes and immediately validates the graph, crashing on error
func MustInitWithScope(registrations []ScopeRegistration) {
	if err := InitWithScope(registrations); err != nil {
		log.Fatalf("Dependency graph initialization failed: %v", err)
	}
}

// RegisterRuntime allows runtime registration of constructors after initialization
func RegisterRuntime(constructor interface{}, scope container.Scope) error {
	if err := dependencyContainer.RegisterRuntimeConstructor(constructor, scope); err != nil {
		return err
	}
	return Validate()
}

// RegisterRuntimeBatch allows runtime registration of multiple constructors after initialization
func RegisterRuntimeBatch(constructors []interface{}, scope container.Scope) error {
	for _, constructor := range constructors {
		if err := dependencyContainer.RegisterRuntimeConstructor(constructor, scope); err != nil {
			return err
		}
	}
	return Validate()
}

// RegisterRuntimeWithScopes registers multiple constructors with individual scopes at runtime
func RegisterRuntimeWithScopes(registrations []ScopeRegistration) error {
	for _, reg := range registrations {
		if err := dependencyContainer.RegisterRuntimeConstructor(reg.Constructor, reg.Scope); err != nil {
			return err
		}
	}
	return Validate()
}

// RegisterNamedConstructor registers a constructor with a specific name and scope
func RegisterNamedConstructor(name string, constructor interface{}, scope container.Scope) error {
	if err := dependencyContainer.RegisterNamedConstructorWithScope(name, constructor, scope); err != nil {
		return err
	}
	return Validate()
}

// Override replaces an existing constructor and clears any cached instances
func Override(constructor interface{}, scope container.Scope) error {
	if err := dependencyContainer.OverrideConstructor(constructor, scope); err != nil {
		return err
	}
	return Validate()
}

// OverrideNamed replaces an existing named constructor and clears any cached instances
func OverrideNamed(name string, constructor interface{}, scope container.Scope) error {
	if err := dependencyContainer.RegisterNamedConstructorWithScope(name, constructor, scope); err != nil {
		return err
	}
	return Validate()
}

// Validate eagerly checks the dependency graph for missing registrations or cycles
func Validate() error {
	return dependencyContainer.Validate()
}

// Reset clears the container state (useful for testing)
func Reset() {
	dependencyContainer = container.New()
}

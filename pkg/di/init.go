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
// @Param constructors []interface{} - list of constructors
func Init(constructors []interface{}) {
	for _, constructor := range constructors {
		if err := dependencyContainer.RegisterConstructor(constructor); err != nil {
			log.Fatalf("Failed to register constructor: %v", err)
		}
	}
}

// InitWithScope registers constructors with specific scopes
// @Param registrations []ScopeRegistration - list of constructor registrations with scopes
func InitWithScope(registrations []ScopeRegistration) {
	for _, reg := range registrations {
		if err := dependencyContainer.RegisterConstructorWithScope(reg.Constructor, reg.Scope); err != nil {
			log.Fatalf("Failed to register constructor: %v", err)
		}
	}
}

// RegisterRuntime allows runtime registration of constructors after initialization
// @Param constructor interface{} - constructor function
// @Param scope container.Scope - lifetime scope for the dependency
func RegisterRuntime(constructor interface{}, scope container.Scope) error {
	return dependencyContainer.RegisterRuntimeConstructor(constructor, scope)
}

// Reset clears the container state (useful for testing)
func Reset() {
	dependencyContainer = container.New()
}

package di

import (
	"fmt"
	"reflect"

	"github.com/binodta/depWeaver/internal/container"
)

// Resolve resolves the instance of the given type from the container
func Resolve[T interface{}]() (T, error) {

	var zero T
	t := reflect.TypeOf(&zero).Elem()
	// Resolve the instance from container
	instance, err := dependencyContainer.Resolve(t)
	if err != nil {
		return zero, fmt.Errorf("failed to resolve type %v: %w", t, err)
	}

	// Try to cast the instance to the desired type
	if instance == nil {
		return zero, fmt.Errorf("resolved instance is nil for type %v", t)
	}

	castedInstance, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("failed to cast resolved instance to type %v", t)
	}

	return castedInstance, nil
}

// ResolveScoped resolves the instance within a specific scope context
// @Param scopeID string - scope context identifier
func ResolveScoped[T interface{}](scopeID string) (T, error) {
	var zero T
	t := reflect.TypeOf(&zero).Elem()
	instance, err := dependencyContainer.ResolveWithScope(t, scopeID)
	if err != nil {
		return zero, fmt.Errorf("failed to resolve type %v in scope %s: %w", t, scopeID, err)
	}

	if instance == nil {
		return zero, fmt.Errorf("resolved instance is nil for type %v", t)
	}

	castedInstance, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("failed to cast resolved instance to type %v", t)
	}

	return castedInstance, nil
}

// GetProvider returns a provider for lazy resolution
// @Param scopeID string - scope context identifier (empty string for default scope)
func GetProvider[T interface{}](scopeID string) container.Provider[T] {
	return container.NewProvider[T](dependencyContainer, scopeID)
}

// CreateScope creates a new scope context and returns its ID
func CreateScope() string {
	return dependencyContainer.CreateScope()
}

// DestroyScope cleans up a scope context and its instances
// @Param scopeID string - scope context identifier to destroy
func DestroyScope(scopeID string) {
	dependencyContainer.DestroyScope(scopeID)
}

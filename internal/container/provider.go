package container

import "reflect"

// Provider provides lazy access to dependencies
type Provider[T any] interface {
	Get() (T, error)
}

type provider[T any] struct {
	container *DependencyContainer
	scopeID   string
}

// Get resolves and returns the dependency
func (p *provider[T]) Get() (T, error) {
	var zero T
	t := reflect.TypeOf(&zero).Elem()
	instance, err := p.container.ResolveWithScope(t, p.scopeID)
	if err != nil {
		return zero, err
	}

	// Type assertion
	result, ok := instance.(T)
	if !ok {
		return zero, nil
	}

	return result, nil
}

// NewProvider creates a provider for lazy resolution
// Note: This is a standalone function (not a method) because Go methods cannot have type parameters
func NewProvider[T any](dc *DependencyContainer, scopeID string) Provider[T] {
	return &provider[T]{
		container: dc,
		scopeID:   scopeID,
	}
}

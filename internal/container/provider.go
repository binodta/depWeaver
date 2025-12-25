package container

import "reflect"

// Provider provides lazy access to dependencies
type Provider[T any] interface {
	Get() (T, error)
}

type provider[T any] struct {
	container *DependencyContainer
	scopeID   string
	name      string
}

// Get resolves and returns the dependency
func (p *provider[T]) Get() (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()

	var instance interface{}
	var err error

	if p.name != "" {
		instance, err = p.container.ResolveNamedWithScope(p.name, t, p.scopeID)
	} else {
		instance, err = p.container.ResolveWithScope(t, p.scopeID)
	}

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
func NewProvider[T any](dc *DependencyContainer, scopeID string, name string) Provider[T] {
	return &provider[T]{
		container: dc,
		scopeID:   scopeID,
		name:      name,
	}
}

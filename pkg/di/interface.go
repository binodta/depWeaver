package di

import (
	"fmt"
	"reflect"
)

// BindInterface binds an interface type to a concrete implementation
// @Param I - interface type
// @Param C - concrete type that implements the interface
func BindInterface[I any, C any]() error {
	interfaceType := reflect.TypeOf((*I)(nil)).Elem()
	concreteType := reflect.TypeOf((*C)(nil)).Elem()

	return dependencyContainer.BindInterface(interfaceType, concreteType)
}

// BindInterfaceNamed binds an interface type to a concrete implementation with a name
// This allows multiple implementations of the same interface
// @Param name - unique name for this binding
// @Param I - interface type
// @Param C - concrete type that implements the interface
func BindInterfaceNamed[I any, C any](name string) error {
	interfaceType := reflect.TypeOf((*I)(nil)).Elem()
	concreteType := reflect.TypeOf((*C)(nil)).Elem()

	return dependencyContainer.BindInterfaceNamed(name, interfaceType, concreteType)
}

// ResolveNamed resolves a dependency by name (for named interface bindings)
// @Param name - name of the binding
// @Param T - type to resolve (typically an interface)
func ResolveNamed[T any](name string) (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()

	instance, err := dependencyContainer.ResolveNamed(name, t)
	if err != nil {
		return zero, fmt.Errorf("failed to resolve named type %v with name %q: %w", t, name, err)
	}

	if instance == nil {
		return zero, fmt.Errorf("resolved instance is nil for type %v with name %q", t, name)
	}

	castedInstance, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("failed to cast resolved instance to type %v", t)
	}

	return castedInstance, nil
}

// ResolveNamedScoped resolves a named dependency within a specific scope
// @Param name - name of the binding
// @Param scopeID - scope context identifier
// @Param T - type to resolve (typically an interface)
func ResolveNamedScoped[T any](name string, scopeID string) (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()

	instance, err := dependencyContainer.ResolveNamedWithScope(name, t, scopeID)
	if err != nil {
		return zero, fmt.Errorf("failed to resolve named type %v with name %q in scope %s: %w", t, name, scopeID, err)
	}

	if instance == nil {
		return zero, fmt.Errorf("resolved instance is nil for type %v with name %q", t, name)
	}

	castedInstance, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("failed to cast resolved instance to type %v", t)
	}

	return castedInstance, nil
}

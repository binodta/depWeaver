package di

import (
	"fmt"
	"reflect"
)

// BindInterface binds an interface type to a concrete implementation
func BindInterface[I any, C any]() error {
	interfaceType := reflect.TypeOf((*I)(nil)).Elem()
	concreteType := reflect.TypeOf((*C)(nil)).Elem()

	if err := dependencyContainer.BindInterface(interfaceType, concreteType); err != nil {
		return err
	}
	return Validate()
}

// BindInterfaceNamed binds an interface type to a concrete implementation with a name
func BindInterfaceNamed[I any, C any](name string) error {
	interfaceType := reflect.TypeOf((*I)(nil)).Elem()
	concreteType := reflect.TypeOf((*C)(nil)).Elem()

	if err := dependencyContainer.BindInterfaceNamed(name, interfaceType, concreteType); err != nil {
		return err
	}
	return Validate()
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

package container

import (
	"fmt"
	"reflect"
)

// ResolveNamed resolves a dependency by name (for named interface bindings)
func (dc *DependencyContainer) ResolveNamed(name string, t reflect.Type) (interface{}, error) {
	return dc.resolveNamedWithScope(name, t, "")
}

// ResolveNamedWithScope resolves a named dependency with a specific scope
func (dc *DependencyContainer) ResolveNamedWithScope(name string, t reflect.Type, scopeID string) (interface{}, error) {
	return dc.resolveNamedWithScope(name, t, scopeID)
}

// resolveNamedWithScope internal method to resolve named dependencies
func (dc *DependencyContainer) resolveNamedWithScope(name string, t reflect.Type, scopeID string) (interface{}, error) {
	// Check if this is an interface type with a named binding
	if t.Kind() == reflect.Interface {
		concreteType, exists := dc.GetNamedInterfaceBinding(name, t)
		if exists {
			// Resolve the concrete type instead
			return dc.resolveWithScope(concreteType, scopeID)
		}
		return nil, fmt.Errorf("no binding found for interface %v with name %q", t, name)
	}

	// For non-interface types, just resolve normally (name is ignored)
	return dc.resolveWithScope(t, scopeID)
}

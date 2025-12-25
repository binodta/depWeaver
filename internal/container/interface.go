package container

import (
	"fmt"
	"reflect"
)

// BindInterface binds an interface type to a concrete implementation
func (dc *DependencyContainer) BindInterface(interfaceType, concreteType reflect.Type) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Validate that interfaceType is actually an interface
	if interfaceType.Kind() != reflect.Interface {
		return fmt.Errorf("type %v is not an interface", interfaceType)
	}

	// Validate that concreteType implements interfaceType
	if !concreteType.Implements(interfaceType) {
		return fmt.Errorf("type %v does not implement interface %v", concreteType, interfaceType)
	}

	// Check if concrete type has a constructor registered
	if _, exists := dc.constructors[concreteType]; !exists {
		return fmt.Errorf("no constructor registered for concrete type %v. Register the constructor first before binding the interface", concreteType)
	}

	// Store the binding
	dc.interfaceBindings[interfaceType] = concreteType
	return nil
}

// BindInterfaceNamed binds an interface type to a concrete implementation with a name
func (dc *DependencyContainer) BindInterfaceNamed(name string, interfaceType, concreteType reflect.Type) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Validate that interfaceType is actually an interface
	if interfaceType.Kind() != reflect.Interface {
		return fmt.Errorf("type %v is not an interface", interfaceType)
	}

	// Validate that concreteType implements interfaceType
	if !concreteType.Implements(interfaceType) {
		return fmt.Errorf("type %v does not implement interface %v", concreteType, interfaceType)
	}

	// Check if concrete type has a constructor registered
	if _, exists := dc.constructors[concreteType]; !exists {
		return fmt.Errorf("no constructor registered for concrete type %v. Register the constructor first before binding the interface", concreteType)
	}

	// Ensure the named bindings map exists for this name
	if dc.namedInterfaceBindings[name] == nil {
		dc.namedInterfaceBindings[name] = make(map[reflect.Type]reflect.Type)
	}

	// Store the named binding
	dc.namedInterfaceBindings[name][interfaceType] = concreteType
	return nil
}

// GetInterfaceBinding returns the concrete type bound to an interface (unnamed)
func (dc *DependencyContainer) GetInterfaceBinding(interfaceType reflect.Type) (reflect.Type, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	concreteType, exists := dc.interfaceBindings[interfaceType]
	return concreteType, exists
}

// GetNamedInterfaceBinding returns the concrete type bound to an interface with a name
func (dc *DependencyContainer) GetNamedInterfaceBinding(name string, interfaceType reflect.Type) (reflect.Type, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	if bindings, exists := dc.namedInterfaceBindings[name]; exists {
		concreteType, found := bindings[interfaceType]
		return concreteType, found
	}

	return nil, false
}

package container

import (
	"fmt"
	"reflect"
)

// ResolveNamed resolves a dependency by name (for named interface bindings)
func (dc *DependencyContainer) ResolveNamed(name string, t reflect.Type) (interface{}, error) {
	return dc.resolveNamedWithScope(name, t, "", nil)
}

// ResolveNamedWithScope resolves a named dependency with a specific scope
func (dc *DependencyContainer) ResolveNamedWithScope(name string, t reflect.Type, scopeID string) (interface{}, error) {
	return dc.resolveNamedWithScope(name, t, scopeID, nil)
}

// resolveNamedWithScope internal method to resolve named dependencies
func (dc *DependencyContainer) resolveNamedWithScope(name string, t reflect.Type, scopeID string, stack []reflect.Type) (interface{}, error) {
	// 1. Check if this is an interface type with a named binding
	if t.Kind() == reflect.Interface {
		concreteType, exists := dc.GetNamedInterfaceBinding(name, t)
		if exists {
			// Resolve the concrete type instead
			return dc.resolveWithScope(concreteType, scopeID, stack)
		}
	}

	// 2. Find the registration in namedConstructors
	dc.mu.RLock()
	nameMap, exists := dc.namedConstructors[name]
	var registration *Registration
	if exists {
		registration, exists = nameMap[t]
	}
	dc.mu.RUnlock()

	if !exists {
		// Fallback: If no named constructor, but it's an interface, return error
		if t.Kind() == reflect.Interface {
			return nil, fmt.Errorf("no binding found for interface %v with name %q", t, name)
		}
		// Fallback: Resolve normally (unnamed)
		return dc.resolveWithScope(t, scopeID, stack)
	}

	// 3. Handle named resolution with separate caches
	switch registration.scope {
	case Singleton:
		return dc.resolveNamedSingleton(name, t, registration, stack)
	case Transient:
		return registration.constructor(dc, scopeID, stack)
	case Scoped:
		return dc.resolveNamedScoped(name, t, registration, scopeID, stack)
	default:
		return nil, fmt.Errorf("unknown scope type for named %v", t)
	}
}

func (dc *DependencyContainer) resolveNamedSingleton(name string, t reflect.Type, registration *Registration, stack []reflect.Type) (interface{}, error) {
	// Fast path
	dc.mu.RLock()
	if typeMap, exists := dc.namedDependencies[name]; exists {
		if dep, exists := typeMap[t]; exists {
			dc.mu.RUnlock()
			return dep, nil
		}
	}
	dc.mu.RUnlock()

	// Slow path: For named singletons, we also use the ResolveNamed entry point
	// which will call resolveNamedWithScope -> resolveNamedSingleton.
	// We need to protect against concurrent creation of named singletons too.
	// For simplicity, we can use the same dc.mu Lock for the whole creation.
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Double-check
	if typeMap, exists := dc.namedDependencies[name]; exists {
		if dep, exists := typeMap[t]; exists {
			return dep, nil
		}
	} else {
		if dc.namedDependencies[name] == nil {
			dc.namedDependencies[name] = make(map[reflect.Type]interface{})
		}
	}

	// Create instance
	instance, err := registration.constructor(dc, "", stack)
	if err != nil {
		return nil, err
	}

	dc.namedDependencies[name][t] = instance
	return instance, nil
}

func (dc *DependencyContainer) resolveNamedScoped(name string, t reflect.Type, registration *Registration, scopeID string, stack []reflect.Type) (interface{}, error) {
	if scopeID == "" {
		return nil, fmt.Errorf("scope ID required for named scoped dependency %v (%s)", t, name)
	}

	dc.mu.RLock()
	if scopeMap, exists := dc.namedScopedInstances[scopeID]; exists {
		if typeMap, exists := scopeMap[name]; exists {
			if dep, exists := typeMap[t]; exists {
				dc.mu.RUnlock()
				return dep, nil
			}
		}
	}
	dc.mu.RUnlock()

	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Ensure maps exist
	if _, exists := dc.namedScopedInstances[scopeID]; !exists {
		dc.namedScopedInstances[scopeID] = make(map[string]map[reflect.Type]interface{})
	}
	if _, exists := dc.namedScopedInstances[scopeID][name]; !exists {
		dc.namedScopedInstances[scopeID][name] = make(map[reflect.Type]interface{})
	}

	// Double-check
	if dep, exists := dc.namedScopedInstances[scopeID][name][t]; exists {
		return dep, nil
	}

	// Create instance
	instance, err := registration.constructor(dc, scopeID, stack)
	if err != nil {
		return nil, err
	}

	dc.namedScopedInstances[scopeID][name][t] = instance
	return instance, nil
}

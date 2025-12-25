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
	// 1. Check if this is an interface type with a named binding
	if t.Kind() == reflect.Interface {
		concreteType, exists := dc.GetNamedInterfaceBinding(name, t)
		if exists {
			// Resolve the concrete type instead
			return dc.resolveWithScope(concreteType, scopeID)
		}
		// If it's an interface but no binding found, we still check namedConstructors
		// maybe someone registered an interface return type directly with a name?
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
		return dc.resolveWithScope(t, scopeID)
	}

	// 3. Handle named resolution with separate caches
	switch registration.scope {
	case Singleton:
		return dc.resolveNamedSingleton(name, t, registration)
	case Transient:
		return registration.constructor(dc, scopeID) // Transients don't need named-specific logic for creation
	case Scoped:
		return dc.resolveNamedScoped(name, t, registration, scopeID)
	default:
		return nil, fmt.Errorf("unknown scope type for named %v", t)
	}
}

func (dc *DependencyContainer) resolveNamedSingleton(name string, t reflect.Type, registration *Registration) (interface{}, error) {
	dc.mu.RLock()
	if typeMap, exists := dc.namedDependencies[name]; exists {
		if dep, exists := typeMap[t]; exists {
			dc.mu.RUnlock()
			return dep, nil
		}
	}
	dc.mu.RUnlock()

	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Double-check
	if typeMap, exists := dc.namedDependencies[name]; exists {
		if dep, exists := typeMap[t]; exists {
			return dep, nil
		}
	} else {
		dc.namedDependencies[name] = make(map[reflect.Type]interface{})
	}

	// Create instance
	instance, err := registration.constructor(dc, "")
	if err != nil {
		return nil, err
	}

	dc.namedDependencies[name][t] = instance
	return instance, nil
}

func (dc *DependencyContainer) resolveNamedScoped(name string, t reflect.Type, registration *Registration, scopeID string) (interface{}, error) {
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
	instance, err := registration.constructor(dc, scopeID)
	if err != nil {
		return nil, err
	}

	dc.namedScopedInstances[scopeID][name][t] = instance
	return instance, nil
}

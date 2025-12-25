package container

import (
	"fmt"
	"reflect"
)

// Resolve public method to resolve dependencies (uses default/empty scope)
// @Param t reflect.Type - type of the dependency
func (dc *DependencyContainer) Resolve(t reflect.Type) (interface{}, error) {
	return dc.resolveWithScope(t, "")
}

// ResolveWithScope public method to resolve dependencies with a specific scope
// @Param t reflect.Type - type of the dependency
// @Param scopeID string - scope context identifier
func (dc *DependencyContainer) ResolveWithScope(t reflect.Type, scopeID string) (interface{}, error) {
	return dc.resolveWithScope(t, scopeID)
}

// resolveWithScope pkg method to resolve dependencies with scope support
// @Param t reflect.Type - type of the dependency
// @Param scopeID string - scope context identifier (empty string for default scope)
// @Return interface{} - instance of the dependency
func (dc *DependencyContainer) resolveWithScope(t reflect.Type, scopeID string) (interface{}, error) {
	// Find the registration for this type
	dc.mu.RLock()
	registration, exists := dc.constructors[t]
	dc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no constructor registered for type %v", t)
	}

	// Handle different scopes
	switch registration.scope {
	case Singleton:
		return dc.resolveSingleton(t, registration, scopeID)
	case Transient:
		return dc.resolveTransient(t, registration, scopeID)
	case Scoped:
		return dc.resolveScoped(t, registration, scopeID)
	default:
		return nil, fmt.Errorf("unknown scope type for %v", t)
	}
}

// resolveSingleton resolves a singleton dependency (created once and cached)
func (dc *DependencyContainer) resolveSingleton(t reflect.Type, registration *Registration, scopeID string) (interface{}, error) {
	// Fast path: try read lock to return already-built singletons
	dc.mu.RLock()
	if dep, exists := dc.dependencies[t]; exists {
		dc.mu.RUnlock()
		return dep, nil
	}
	dc.mu.RUnlock()

	// Slow path: acquire write lock to safely check/create
	dc.mu.Lock()

	// Double-check after acquiring the write lock
	if dep, exists := dc.dependencies[t]; exists {
		dc.mu.Unlock()
		return dep, nil
	}

	// Check for circular dependencies with detailed error reporting
	if dc.creating[t] {
		dc.mu.Unlock()
		return nil, fmt.Errorf("circular dependency detected: %s", dc.formatDependencyChain(t))
	}

	// Mark as currently being created to prevent recursion
	dc.creating[t] = true
	dc.resolutionStack = append(dc.resolutionStack, t)
	dc.mu.Unlock()

	// Create the instance
	instance, err := registration.constructor(dc, scopeID)

	// Clean up resolution tracking
	dc.mu.Lock()
	delete(dc.creating, t)
	if len(dc.resolutionStack) > 0 {
		dc.resolutionStack = dc.resolutionStack[:len(dc.resolutionStack)-1]
	}

	if err != nil {
		dc.mu.Unlock()
		return nil, err
	}

	// Store the created instance
	dc.dependencies[t] = instance
	dc.mu.Unlock()

	return instance, nil
}

// resolveTransient resolves a transient dependency (created every time)
func (dc *DependencyContainer) resolveTransient(t reflect.Type, registration *Registration, scopeID string) (interface{}, error) {
	// Check for circular dependencies
	dc.mu.Lock()
	if dc.creating[t] {
		dc.mu.Unlock()
		return nil, fmt.Errorf("circular dependency detected: %s", dc.formatDependencyChain(t))
	}

	// Mark as currently being created to prevent recursion
	dc.creating[t] = true
	dc.resolutionStack = append(dc.resolutionStack, t)
	dc.mu.Unlock()

	// Create the instance (always new)
	instance, err := registration.constructor(dc, scopeID)

	// Clean up resolution tracking
	dc.mu.Lock()
	delete(dc.creating, t)
	if len(dc.resolutionStack) > 0 {
		dc.resolutionStack = dc.resolutionStack[:len(dc.resolutionStack)-1]
	}
	dc.mu.Unlock()

	if err != nil {
		return nil, err
	}

	return instance, nil
}

// resolveScoped resolves a scoped dependency (created once per scope context)
func (dc *DependencyContainer) resolveScoped(t reflect.Type, registration *Registration, scopeID string) (interface{}, error) {
	if scopeID == "" {
		return nil, fmt.Errorf("scope ID required for scoped dependency %v", t)
	}

	// Fast path: check if already created in this scope
	dc.mu.RLock()
	if scopeCache, exists := dc.scopedInstances[scopeID]; exists {
		if dep, exists := scopeCache[t]; exists {
			dc.mu.RUnlock()
			return dep, nil
		}
	}
	dc.mu.RUnlock()

	// Slow path: create the instance
	dc.mu.Lock()

	// Double-check after acquiring write lock
	if scopeCache, exists := dc.scopedInstances[scopeID]; exists {
		if dep, exists := scopeCache[t]; exists {
			dc.mu.Unlock()
			return dep, nil
		}
	}

	// Ensure scope exists
	if _, exists := dc.scopedInstances[scopeID]; !exists {
		dc.scopedInstances[scopeID] = make(map[reflect.Type]interface{})
	}

	// Check for circular dependencies
	if dc.creating[t] {
		dc.mu.Unlock()
		return nil, fmt.Errorf("circular dependency detected: %s", dc.formatDependencyChain(t))
	}

	// Mark as currently being created
	dc.creating[t] = true
	dc.resolutionStack = append(dc.resolutionStack, t)
	dc.mu.Unlock()

	// Create the instance
	instance, err := registration.constructor(dc, scopeID)

	// Clean up resolution tracking
	dc.mu.Lock()
	delete(dc.creating, t)
	if len(dc.resolutionStack) > 0 {
		dc.resolutionStack = dc.resolutionStack[:len(dc.resolutionStack)-1]
	}

	if err != nil {
		dc.mu.Unlock()
		return nil, err
	}

	// Store in scope cache
	dc.scopedInstances[scopeID][t] = instance
	dc.mu.Unlock()

	return instance, nil
}

// formatDependencyChain creates a readable string showing the circular dependency path
// Note: This method should be called while holding the lock
func (dc *DependencyContainer) formatDependencyChain(circularType reflect.Type) string {
	if len(dc.resolutionStack) == 0 {
		return fmt.Sprintf("%v -> %v (circular)", circularType, circularType)
	}

	chain := ""
	for i, t := range dc.resolutionStack {
		if i > 0 {
			chain += " -> "
		}
		chain += t.String()
	}
	chain += " -> " + circularType.String()

	return chain
}

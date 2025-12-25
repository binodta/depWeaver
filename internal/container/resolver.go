package container

import (
	"fmt"
	"reflect"
)

// Resolve public method to resolve dependencies (uses default/empty scope)
func (dc *DependencyContainer) Resolve(t reflect.Type) (interface{}, error) {
	return dc.resolveWithScope(t, "", nil)
}

// ResolveWithScope public method to resolve dependencies with a specific scope
func (dc *DependencyContainer) ResolveWithScope(t reflect.Type, scopeID string) (interface{}, error) {
	return dc.resolveWithScope(t, scopeID, nil)
}

// resolveWithScope pkg method to resolve dependencies with scope support
// @Param stack []reflect.Type - Call stack for the CURRENT resolution chain (local to goroutine)
func (dc *DependencyContainer) resolveWithScope(t reflect.Type, scopeID string, stack []reflect.Type) (interface{}, error) {
	// Check if this is an interface type with a binding
	if t.Kind() == reflect.Interface {
		concreteType, exists := dc.GetInterfaceBinding(t)
		if exists {
			return dc.resolveWithScope(concreteType, scopeID, stack)
		}
	}

	// 1. Check for circular dependencies in the CURRENT call stack
	for _, stackType := range stack {
		if stackType == t {
			return nil, fmt.Errorf("circular dependency detected: %s", dc.formatDependencyChain(t, stack))
		}
	}

	// 2. Find the registration for this type
	dc.mu.RLock()
	registration, exists := dc.constructors[t]
	dc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no constructor registered for type %v", t)
	}

	// Update stack
	newStack := append(stack, t)

	// 3. Handle different scopes
	switch registration.scope {
	case Singleton:
		return dc.resolveSingleton(t, registration, scopeID, newStack)
	case Transient:
		return dc.resolveTransient(t, registration, scopeID, newStack)
	case Scoped:
		return dc.resolveScoped(t, registration, scopeID, newStack)
	default:
		return nil, fmt.Errorf("unknown scope type for %v", t)
	}
}

// resolveSingleton resolves a singleton dependency (created once and cached)
func (dc *DependencyContainer) resolveSingleton(t reflect.Type, registration *Registration, scopeID string, stack []reflect.Type) (interface{}, error) {
	// 1. Fast path: read lock
	dc.mu.RLock()
	if dep, exists := dc.dependencies[t]; exists {
		dc.mu.RUnlock()
		return dep, nil
	}

	// 2. Check if another goroutine is already building this
	waitChan, inProg := dc.inProgress[t]
	dc.mu.RUnlock()

	if inProg {
		<-waitChan           // Wait for the builder to finish
		return dc.Resolve(t) // Recursive call (will hit fast path)
	}

	// 3. Slow path: become the builder
	dc.mu.Lock()
	// Double check
	if dep, exists := dc.dependencies[t]; exists {
		dc.mu.Unlock()
		return dep, nil
	}
	if waitChan, inProg = dc.inProgress[t]; inProg {
		dc.mu.Unlock()
		<-waitChan
		return dc.Resolve(t)
	}

	// Mark as in-progress
	done := make(chan struct{})
	dc.inProgress[t] = done
	dc.mu.Unlock()

	// Ensure we close the channel and cleanup even if constructor panics
	defer func() {
		dc.mu.Lock()
		delete(dc.inProgress, t)
		close(done)
		dc.mu.Unlock()
	}()

	// Create the instance
	instance, err := registration.constructor(dc, scopeID, stack)
	if err != nil {
		return nil, err
	}

	// Store the created instance
	dc.mu.Lock()
	dc.dependencies[t] = instance
	dc.mu.Unlock()

	return instance, nil
}

// resolveTransient resolves a transient dependency (created every time)
func (dc *DependencyContainer) resolveTransient(t reflect.Type, registration *Registration, scopeID string, stack []reflect.Type) (interface{}, error) {
	// Create the instance (no caching needed, cycle detection already done in resolveWithScope)
	return registration.constructor(dc, scopeID, stack)
}

// resolveScoped resolves a scoped dependency (created once per scope context)
func (dc *DependencyContainer) resolveScoped(t reflect.Type, registration *Registration, scopeID string, stack []reflect.Type) (interface{}, error) {
	if scopeID == "" {
		return nil, fmt.Errorf("scope ID required for scoped dependency %v", t)
	}

	dc.mu.RLock()
	if scopeCache, exists := dc.scopedInstances[scopeID]; exists {
		if dep, exists := scopeCache[t]; exists {
			dc.mu.RUnlock()
			return dep, nil
		}
	}
	dc.mu.RUnlock()

	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Double check
	if scopeCache, exists := dc.scopedInstances[scopeID]; exists {
		if dep, exists := scopeCache[t]; exists {
			return dep, nil
		}
	} else {
		dc.scopedInstances[scopeID] = make(map[reflect.Type]interface{})
	}

	// Create the instance
	// Note: Scoped creators don't wait for each other across different scopeIDs.
	// Within the same scopeID, they are protected by dc.mu.
	instance, err := registration.constructor(dc, scopeID, stack)
	if err != nil {
		return nil, err
	}

	dc.scopedInstances[scopeID][t] = instance
	return instance, nil
}

func (dc *DependencyContainer) formatDependencyChain(circularType reflect.Type, stack []reflect.Type) string {
	if len(stack) == 0 {
		return fmt.Sprintf("%v -> %v (circular)", circularType, circularType)
	}

	chain := ""
	for i, t := range stack {
		if i > 0 {
			chain += " -> "
		}
		chain += t.String()
	}
	chain += " -> " + circularType.String()

	return chain
}

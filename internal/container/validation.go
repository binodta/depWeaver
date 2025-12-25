package container

import (
	"fmt"
	"reflect"
)

type nodeKey struct {
	t    reflect.Type
	name string
}

// Validate eagerly checks the entire dependency graph for missing dependencies and circular dependencies
func (dc *DependencyContainer) Validate() error {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	visited := make(map[nodeKey]bool)

	// Check unnamed constructors
	for t := range dc.constructors {
		if err := dc.validateNode(nodeKey{t: t}, visited, make(map[nodeKey]bool), nil); err != nil {
			return err
		}
	}

	// Check named constructors
	for name, nameMap := range dc.namedConstructors {
		for t := range nameMap {
			if err := dc.validateNode(nodeKey{t: t, name: name}, visited, make(map[nodeKey]bool), nil); err != nil {
				return err
			}
		}
	}

	return nil
}

func (dc *DependencyContainer) validateNode(key nodeKey, visited, inProgress map[nodeKey]bool, stack []nodeKey) error {
	if inProgress[key] {
		chain := ""
		for _, node := range stack {
			if node.name != "" {
				chain += fmt.Sprintf("[%s]%v -> ", node.name, node.t)
			} else {
				chain += fmt.Sprintf("%v -> ", node.t)
			}
		}
		if key.name != "" {
			chain += fmt.Sprintf("[%s]%v", key.name, key.t)
		} else {
			chain += fmt.Sprintf("%v", key.t)
		}
		return fmt.Errorf("circular dependency detected: %s", chain)
	}

	if visited[key] {
		return nil
	}

	inProgress[key] = true
	newStack := append(stack, key)
	defer func() {
		inProgress[key] = false
		visited[key] = true
	}()

	var reg *Registration
	var exists bool
	t := key.t

	if key.name != "" {
		// Named resolution
		if t.Kind() == reflect.Interface {
			concreteType, ok := dc.GetNamedInterfaceBinding(key.name, t)
			if ok {
				return dc.validateNode(nodeKey{t: concreteType}, visited, inProgress, newStack)
			}
		}
		nameMap, ok := dc.namedConstructors[key.name]
		if ok {
			reg, exists = nameMap[t]
		}
	} else {
		// Unnamed resolution
		if t.Kind() == reflect.Interface {
			concreteType, ok := dc.GetInterfaceBinding(t)
			if ok {
				return dc.validateNode(nodeKey{t: concreteType}, visited, inProgress, newStack)
			}
			return fmt.Errorf("no binding found for interface %v", t)
		}
		reg, exists = dc.constructors[t]
	}

	if !exists {
		if key.name != "" {
			return fmt.Errorf("no constructor found for named dependency %v (%s)", t, key.name)
		}
		return fmt.Errorf("no constructor registered for type %v", t)
	}

	// Check dependencies (all parameters are currently resolved UNNAMED)
	for _, paramType := range reg.paramTypes {
		if err := dc.validateNode(nodeKey{t: paramType}, visited, inProgress, newStack); err != nil {
			return err
		}
	}

	return nil
}

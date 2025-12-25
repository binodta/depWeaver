package container

import (
	"crypto/rand"
	"encoding/hex"
	"reflect"
)

// CreateScope creates a new scope context and returns its ID
func (dc *DependencyContainer) CreateScope() string {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	scopeID := generateScopeID()
	dc.scopedInstances[scopeID] = make(map[reflect.Type]interface{})
	return scopeID
}

// DestroyScope removes a scope and its instances (including named ones)
func (dc *DependencyContainer) DestroyScope(scopeID string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	delete(dc.scopedInstances, scopeID)
	delete(dc.namedScopedInstances, scopeID)
}

// DestroyAllScopes removes all active scope contexts and their instances
func (dc *DependencyContainer) DestroyAllScopes() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.scopedInstances = make(map[string]map[reflect.Type]interface{})
	dc.namedScopedInstances = make(map[string]map[string]map[reflect.Type]interface{})
}

// generateScopeID generates a unique scope identifier
func generateScopeID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple counter-based approach if random fails
		// In production, you might want to handle this differently
		panic("failed to generate scope ID: " + err.Error())
	}
	return hex.EncodeToString(bytes)
}

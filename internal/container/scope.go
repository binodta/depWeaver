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

// DestroyScope removes a scope and its instances
func (dc *DependencyContainer) DestroyScope(scopeID string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	delete(dc.scopedInstances, scopeID)
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

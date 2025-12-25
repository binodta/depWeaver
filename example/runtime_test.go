package main

import (
	"fmt"
	"testing"

	"github.com/binodta/depWeaver/internal/container"
	"github.com/binodta/depWeaver/pkg/di"
)

// Test types for scope testing
type Counter struct {
	count int
}

func NewCounter() *Counter {
	return &Counter{count: 0}
}

type Database struct {
	connectionID string
}

var dbConnectionCounter = 0

func NewDatabase() *Database {
	dbConnectionCounter++
	return &Database{connectionID: fmt.Sprintf("conn-%d", dbConnectionCounter)}
}

type RequestContext struct {
	requestID string
}

func NewRequestContext() *RequestContext {
	return &RequestContext{requestID: "req-123"}
}

// TestSingletonScope verifies that Singleton dependencies are created once and cached
func TestSingletonScope(t *testing.T) {
	di.Reset()
	dbConnectionCounter = 0

	registrations := []di.ScopeRegistration{
		{Constructor: NewDatabase, Scope: container.Singleton},
	}

	di.InitWithScope(registrations)

	db1, err := di.Resolve[*Database]()
	if err != nil {
		t.Fatalf("Failed to resolve Database: %v", err)
	}

	db2, err := di.Resolve[*Database]()
	if err != nil {
		t.Fatalf("Failed to resolve Database: %v", err)
	}

	// Should be the same instance
	if db1 != db2 {
		t.Error("Expected same instance for Singleton scope")
	}

	if db1.connectionID != db2.connectionID {
		t.Errorf("Expected same connection ID, got %s and %s", db1.connectionID, db2.connectionID)
	}

	if dbConnectionCounter != 1 {
		t.Errorf("Expected constructor to be called once, but was called %d times", dbConnectionCounter)
	}

	t.Logf("Singleton test passed: same instance %s", db1.connectionID)
}

// TestTransientScope verifies that Transient dependencies are created every time
func TestTransientScope(t *testing.T) {
	di.Reset()
	dbConnectionCounter = 0

	registrations := []di.ScopeRegistration{
		{Constructor: NewDatabase, Scope: container.Transient},
	}

	di.InitWithScope(registrations)

	db1, err := di.Resolve[*Database]()
	if err != nil {
		t.Fatalf("Failed to resolve Database: %v", err)
	}

	db2, err := di.Resolve[*Database]()
	if err != nil {
		t.Fatalf("Failed to resolve Database: %v", err)
	}

	// Should be different instances
	if db1 == db2 {
		t.Error("Expected different instances for Transient scope")
	}

	if db1.connectionID == db2.connectionID {
		t.Errorf("Expected different connection IDs, got %s and %s", db1.connectionID, db2.connectionID)
	}

	if dbConnectionCounter != 2 {
		t.Errorf("Expected constructor to be called twice, but was called %d times", dbConnectionCounter)
	}

	t.Logf("Transient test passed: different instances %s and %s", db1.connectionID, db2.connectionID)
}

// TestScopedScope verifies that Scoped dependencies are created once per scope
func TestScopedScope(t *testing.T) {
	di.Reset()
	registrations := []di.ScopeRegistration{
		{Constructor: NewRequestContext, Scope: container.Scoped},
	}

	di.InitWithScope(registrations)

	// Create first scope
	scope1 := di.CreateScope()
	defer di.DestroyScope(scope1)

	ctx1a, err := di.ResolveScoped[*RequestContext](scope1)
	if err != nil {
		t.Fatalf("Failed to resolve RequestContext in scope1: %v", err)
	}

	ctx1b, err := di.ResolveScoped[*RequestContext](scope1)
	if err != nil {
		t.Fatalf("Failed to resolve RequestContext in scope1 again: %v", err)
	}

	// Should be the same instance within the same scope
	if ctx1a != ctx1b {
		t.Error("Expected same instance within the same scope")
	}

	// Create second scope
	scope2 := di.CreateScope()
	defer di.DestroyScope(scope2)

	ctx2, err := di.ResolveScoped[*RequestContext](scope2)
	if err != nil {
		t.Fatalf("Failed to resolve RequestContext in scope2: %v", err)
	}

	// Should be different instances across different scopes
	if ctx1a == ctx2 {
		t.Error("Expected different instances across different scopes")
	}

	t.Logf("Scoped test passed: same instance within scope, different across scopes")
}

// TestScopedWithoutScopeID verifies that scoped dependencies require a scope ID
func TestScopedWithoutScopeID(t *testing.T) {
	di.Reset()
	registrations := []di.ScopeRegistration{
		{Constructor: NewRequestContext, Scope: container.Scoped},
	}

	di.InitWithScope(registrations)

	_, err := di.Resolve[*RequestContext]()
	if err == nil {
		t.Fatal("Expected error when resolving scoped dependency without scope ID")
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}

	t.Logf("Scoped without scope ID test passed: %v", err)
}

// TestProviderLazyLoading verifies that providers defer dependency creation
func TestProviderLazyLoading(t *testing.T) {
	di.Reset()
	dbConnectionCounter = 0

	registrations := []di.ScopeRegistration{
		{Constructor: NewDatabase, Scope: container.Singleton},
	}

	di.InitWithScope(registrations)

	// Get provider - should not create the instance yet
	provider := di.GetProvider[*Database]("")

	if dbConnectionCounter != 0 {
		t.Errorf("Expected constructor not to be called yet, but was called %d times", dbConnectionCounter)
	}

	// Now actually get the instance
	db, err := provider.Get()
	if err != nil {
		t.Fatalf("Failed to get Database from provider: %v", err)
	}

	if dbConnectionCounter != 1 {
		t.Errorf("Expected constructor to be called once, but was called %d times", dbConnectionCounter)
	}

	// Get again - should return cached instance
	db2, err := provider.Get()
	if err != nil {
		t.Fatalf("Failed to get Database from provider again: %v", err)
	}

	if db != db2 {
		t.Error("Expected same instance from provider")
	}

	if dbConnectionCounter != 1 {
		t.Errorf("Expected constructor to still be called only once, but was called %d times", dbConnectionCounter)
	}

	t.Logf("Provider lazy loading test passed: instance created on first Get()")
}

// TestRuntimeRegistration verifies that dependencies can be registered after Init
func TestRuntimeRegistration(t *testing.T) {
	di.Reset()
	// Initialize with base dependencies
	di.Init([]interface{}{NewCounter})

	// Resolve base dependency
	counter, err := di.Resolve[*Counter]()
	if err != nil {
		t.Fatalf("Failed to resolve Counter: %v", err)
	}

	if counter == nil {
		t.Fatal("Expected non-nil Counter")
	}

	// Now register a new dependency at runtime
	dbConnectionCounter = 0
	err = di.RegisterRuntime(NewDatabase, container.Transient)
	if err != nil {
		t.Fatalf("Failed to register Database at runtime: %v", err)
	}

	// Resolve the runtime-registered dependency
	db, err := di.Resolve[*Database]()
	if err != nil {
		t.Fatalf("Failed to resolve runtime-registered Database: %v", err)
	}

	if db == nil {
		t.Fatal("Expected non-nil Database")
	}

	t.Logf("Runtime registration test passed: registered and resolved Database at runtime")
}

// TestMixedScopes verifies that different scopes can coexist
func TestMixedScopes(t *testing.T) {
	di.Reset()
	dbConnectionCounter = 0

	type Service struct {
		db      *Database
		counter *Counter
	}

	NewService := func(db *Database, counter *Counter) *Service {
		return &Service{db: db, counter: counter}
	}

	registrations := []di.ScopeRegistration{
		{Constructor: NewDatabase, Scope: container.Transient}, // New DB each time
		{Constructor: NewCounter, Scope: container.Singleton},  // Same counter always
		{Constructor: NewService, Scope: container.Singleton},  // Service is singleton
	}

	di.InitWithScope(registrations)

	svc1, err := di.Resolve[*Service]()
	if err != nil {
		t.Fatalf("Failed to resolve Service: %v", err)
	}

	svc2, err := di.Resolve[*Service]()
	if err != nil {
		t.Fatalf("Failed to resolve Service again: %v", err)
	}

	// Services should be the same (singleton)
	if svc1 != svc2 {
		t.Error("Expected same Service instance (singleton)")
	}

	// Counters should be the same (singleton)
	if svc1.counter != svc2.counter {
		t.Error("Expected same Counter instance (singleton)")
	}

	// But the service was only created once, so DB should be the same
	if svc1.db != svc2.db {
		t.Error("Expected same DB instance (service created once)")
	}

	t.Logf("Mixed scopes test passed")
}

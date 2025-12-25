package main

import (
	"testing"

	"github.com/binodta/depWeaver/internal/container"
	"github.com/binodta/depWeaver/pkg/di"
)

// Define interfaces
type ITestLogger interface {
	Log(msg string)
}

type ITestCache interface {
	Get(key string) (string, error)
	Set(key, value string) error
}

type ITestRepository interface {
	Save(data string) error
}

type ITestDatabase interface {
	Execute(query string) error
}

// Concrete implementations
type ConsoleLogger struct {
	prefix string
}

func (l *ConsoleLogger) Log(msg string) {
	// Implementation
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{prefix: "[LOG]"}
}

type RedisCache struct {
	id int
}

func (c *RedisCache) Get(key string) (string, error) {
	return "", nil
}

func (c *RedisCache) Set(key, value string) error {
	return nil
}

func NewRedisCache() *RedisCache {
	return &RedisCache{}
}

type MemoryCache struct{}

func (c *MemoryCache) Get(key string) (string, error) {
	return "", nil
}

func (c *MemoryCache) Set(key, value string) error {
	return nil
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{}
}

type SQLDatabase struct{}

func (db *SQLDatabase) Execute(query string) error {
	return nil
}

func NewSQLDatabase() *SQLDatabase {
	return &SQLDatabase{}
}

type SQLRepository struct {
	db ITestDatabase
}

func (r *SQLRepository) Save(data string) error {
	return r.db.Execute("INSERT...")
}

func NewSQLRepository(db ITestDatabase) *SQLRepository {
	return &SQLRepository{db: db}
}

// TestBasicInterfaceBinding tests basic interface to concrete type binding
func TestBasicInterfaceBinding(t *testing.T) {
	di.Reset()

	// Register constructor
	di.Init([]interface{}{NewConsoleLogger})

	// Bind interface
	err := di.BindInterface[ITestLogger, *ConsoleLogger]()
	if err != nil {
		t.Fatalf("Failed to bind interface: %v", err)
	}

	// Resolve by interface
	logger, err := di.Resolve[ITestLogger]()
	if err != nil {
		t.Fatalf("Failed to resolve ITestLogger interface: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected non-nil logger")
	}

	// Verify it's the same instance as concrete resolution (Singleton)
	concreteLogger, _ := di.Resolve[*ConsoleLogger]()
	if logger != concreteLogger {
		t.Error("Expected interface and concrete resolution to return same instance")
	}

	t.Log("Basic interface binding test passed")
}

// TestNamedInterfaceBinding tests multiple implementations with named bindings
func TestNamedInterfaceBinding(t *testing.T) {
	di.Reset()

	// Register constructors
	di.Init([]interface{}{NewRedisCache, NewMemoryCache})

	// Bind with names
	err := di.BindInterfaceNamed[ITestCache, *RedisCache]("redis")
	if err != nil {
		t.Fatalf("Failed to bind redis cache: %v", err)
	}

	err = di.BindInterfaceNamed[ITestCache, *MemoryCache]("memory")
	if err != nil {
		t.Fatalf("Failed to bind memory cache: %v", err)
	}

	// Resolve by name
	redisCache, err := di.ResolveNamed[ITestCache]("redis")
	if err != nil {
		t.Fatalf("Failed to resolve redis cache: %v", err)
	}

	memCache, err := di.ResolveNamed[ITestCache]("memory")
	if err != nil {
		t.Fatalf("Failed to resolve memory cache: %v", err)
	}

	if redisCache == nil || memCache == nil {
		t.Fatal("Expected non-nil cache instances")
	}

	// Verify they are different instances
	if redisCache == memCache {
		t.Error("Expected different instances for different named bindings")
	}

	t.Log("Named interface binding test passed")
}

// TestInterfaceWithDependencies tests interface binding with dependency injection
func TestInterfaceWithDependencies(t *testing.T) {
	di.Reset()

	// Register constructors
	di.Init([]interface{}{NewSQLDatabase, NewSQLRepository})

	// Bind interfaces
	err := di.BindInterface[ITestDatabase, *SQLDatabase]()
	if err != nil {
		t.Fatalf("Failed to bind ITestDatabase interface: %v", err)
	}

	err = di.BindInterface[ITestRepository, *SQLRepository]()
	if err != nil {
		t.Fatalf("Failed to bind ITestRepository interface: %v", err)
	}

	// Resolve repository (should automatically resolve ITestDatabase dependency)
	repo, err := di.Resolve[ITestRepository]()
	if err != nil {
		t.Fatalf("Failed to resolve ITestRepository: %v", err)
	}

	if repo == nil {
		t.Fatal("Expected non-nil repository")
	}

	// Verify the repository has the database dependency
	sqlRepo, ok := repo.(*SQLRepository)
	if !ok {
		t.Fatal("Expected *SQLRepository type")
	}

	if sqlRepo.db == nil {
		t.Error("Expected repository to have database dependency")
	}

	t.Log("Interface with dependencies test passed")
}

// TestInterfaceBindingErrors tests error cases
func TestInterfaceBindingErrors(t *testing.T) {
	di.Reset()

	// Test 1: Binding non-interface type
	type NotAnInterface struct{}
	err := di.BindInterface[NotAnInterface, *ConsoleLogger]()
	if err == nil {
		t.Error("Expected error when binding non-interface type")
	}
	t.Logf("Error for non-interface: %v", err)

	// Test 2: Binding to type that doesn't implement interface
	di.Init([]interface{}{NewConsoleLogger})
	err = di.BindInterface[ITestCache, *ConsoleLogger]()
	if err == nil {
		t.Error("Expected error when concrete type doesn't implement interface")
	}
	t.Logf("Error for non-implementation: %v", err)

	// Test 3: Binding before registering constructor
	di.Reset()
	err = di.BindInterface[ITestLogger, *ConsoleLogger]()
	if err == nil {
		t.Error("Expected error when constructor not registered")
	}
	t.Logf("Error for missing constructor: %v", err)

	// Test 4: Resolving named binding that doesn't exist
	di.Reset()
	di.Init([]interface{}{NewRedisCache})
	di.BindInterfaceNamed[ITestCache, *RedisCache]("redis")

	_, err = di.ResolveNamed[ITestCache]("nonexistent")
	if err == nil {
		t.Error("Expected error when resolving non-existent named binding")
	}
	t.Logf("Error for non-existent binding: %v", err)

	t.Log("Interface binding error tests passed")
}

// TestInterfaceWithScopes tests interface binding with different scopes
func TestInterfaceWithScopes(t *testing.T) {
	di.Reset()

	// Register with different scopes
	registrations := []di.ScopeRegistration{
		{Constructor: NewConsoleLogger, Scope: container.Singleton},
		{Constructor: NewRedisCache, Scope: container.Transient},
	}

	di.InitWithScope(registrations)

	// Bind interfaces
	di.BindInterface[ITestLogger, *ConsoleLogger]()
	di.BindInterface[ITestCache, *RedisCache]()

	// Test Singleton behavior
	logger1, _ := di.Resolve[ITestLogger]()
	logger2, _ := di.Resolve[ITestLogger]()
	if logger1 != logger2 {
		t.Error("Expected same instance for Singleton interface binding")
	}

	// Test Transient behavior
	cache1, _ := di.Resolve[ITestCache]()
	cache2, _ := di.Resolve[ITestCache]()
	t.Logf("Cache 1: %p (%T)", cache1, cache1)
	t.Logf("Cache 2: %p (%T)", cache2, cache2)
	if cache1 == cache2 {
		t.Error("Expected different instances for Transient interface binding")
	}

	t.Log("Interface with scopes test passed")
}

# DepWeaver

Lightweight, reflection-powered dependency injection for Go with runtime resolution capabilities. Define constructors for your types and let DepWeaver build the dependency graph for you at runtime.

## Features

- **Minimal API surface** - Simple, intuitive API
- **Constructor-based injection** - Order independent registration
- **Flexible return types** - Supports `T` or `(T, error)` signatures
- **Scope management** - Singleton, Transient, and Scoped lifetimes
- **Interface-based resolution** - Bind interfaces to concrete implementations
- **Named dependencies** - Support for multiple implementations and concrete instances
- **Lazy loading** - Provider pattern for deferred dependency creation
- **Eager validation** - Graph validation for cycles and missing dependencies
- **Test Overrides** - Easily swap implementations for mocking in tests
- **Runtime registration** - Register dependencies dynamically
- **Circular dependency detection** - Clear error messages with dependency chains
- **Thread-safe** - Concurrent resolution and registration
- **100% backward compatible** - Existing code works unchanged

Installation

Use the Go tool:

```
go get github.com/binodta/depWeaver
```

## Quick Start

### Basic Usage (Singleton - Default)

```go
package main

import (
    "fmt"
    "github.com/binodta/depWeaver/pkg/di"
)

// Types
type Config struct{ DSN string }
type DB struct{ DSN string }
type Repo struct{ DB *DB }
type Service struct{ Repo *Repo }

// Constructors (order does not matter)
func NewConfig() *Config { return &Config{DSN: "postgres://..."} }
func NewDB(c *Config) (*DB, error) { return &DB{DSN: c.DSN}, nil }
func NewRepo(db *DB) *Repo { return &Repo{DB: db} }
func NewService(r *Repo) *Service { return &Service{Repo: r} }

func main() {
    // All dependencies default to Singleton scope
    di.Init([]interface{}{
        NewService,
        NewConfig,
        NewDB,          // (T, error) supported
        NewRepo,
    })

    svc, err := di.Resolve[*Service]()
    if err != nil {
        panic(err)
    }
    fmt.Println("Service ready with DSN:", svc.Repo.DB.DSN)
}
```

### Advanced Usage (With Scopes)

```go
import (
    "github.com/binodta/depWeaver/pkg/di"
    "github.com/binodta/depWeaver/internal/container"
)

func main() {
    // Register with different scopes
    registrations := []di.ScopeRegistration{
        {Constructor: NewConfig, Scope: container.Singleton},   // Created once
        {Constructor: NewDB, Scope: container.Transient},       // New each time
        {Constructor: NewRequestCtx, Scope: container.Scoped},  // Once per scope
    }
    
    di.InitWithScope(registrations)
    
    // Singleton resolution
    config, _ := di.Resolve[*Config]()
    
    // Transient resolution (new instance each time)
    db1, _ := di.Resolve[*DB]()
    db2, _ := di.Resolve[*DB]()  // db1 != db2
    
    // Scoped resolution (per-request, per-session, etc.)
    scopeID := di.CreateScope()
    defer di.DestroyScope(scopeID)
    
    ctx1, _ := di.ResolveScoped[*RequestCtx](scopeID)
    ctx2, _ := di.ResolveScoped[*RequestCtx](scopeID)  // ctx1 == ctx2
}
```

## API Overview

### Basic API (Backward Compatible)

**`di.Init(constructors []interface{}) error`**
- Register components with Singleton scope
- Automatically runs `di.Validate()` and returns error on failure

**`di.MustInit(constructors []interface{})`**
- Same as `Init` but crashes with `log.Fatalf` on error (ideal for app startup)

**`di.Resolve[T]() (T, error)`**
- Resolve and return an instance of type `T`
- Dependencies are resolved automatically
- Singleton instances are cached

### Scope Management API

**`di.InitWithScope(registrations []ScopeRegistration) error`**
- Register constructors with specific scopes
- Automatically runs `di.Validate()`

**`di.MustInitWithScope(registrations []ScopeRegistration)`**
- Same as `InitWithScope` but crashes on error
- Scopes: `container.Singleton`, `container.Transient`, `container.Scoped`

**`di.ResolveScoped[T](scopeID string) (T, error)`**
- Resolve instance within a specific scope context
- Required for `Scoped` dependencies
- Returns cached instance within the same scope

**`di.CreateScope() string`**
- Create a new scope context
- Returns unique scope ID
- Use with `defer di.DestroyScope(scopeID)` for cleanup

**`di.DestroyScope(scopeID string)`**
- Clean up scope and its cached instances
- Should be called when scope is no longer needed

**`di.DestroyAllScopes()`**
- Force cleanup of ALL active scopes (both unnamed and named)
- Useful as a contingency for memory leaks or application teardown

**`di.Validate() error`**
- Eagerly check the entire dependency graph for cycles and missing registrations
- Recommended to call during application startup

**`di.Override(constructor interface{}, scope container.Scope) error`**
- Replace an existing registration and clear its cache
- Ideal for injecting mocks/stubs during testing

### Interface Binding API

**`di.BindInterface[I any, C any]() error`**
- Bind interface type `I` to concrete type `C`
- `C` must have a registered constructor
- `C` must implement `I`

**`di.BindInterfaceNamed[I any, C any](name string) error`**
- Bind interface type `I` to concrete type `C` with a unique name
- Allows multiple implementations of the same interface

**`di.ResolveNamed[T any](name string) (T, error)`**
- Resolve a named binding (interface or concrete)
- Used when multiple implementations of the same type exist

**`di.ResolveNamedScoped[T any](name string, scopeID string) (T, error)`**
- Resolve a named binding within a specific scope

**`di.RegisterNamedConstructor(name string, constructor interface{}, scope container.Scope) error`**
- Register a concrete type with a unique name
- Allows multiple instances of the same concrete type with independent caching

### Lazy Loading API

**`di.GetProvider[T](scopeID string) container.Provider[T]`**
- Get a provider for lazy dependency resolution
- Dependency is not created until `provider.Get()` is called
- Use empty string `""` for default scope

**`provider.Get() (T, error)`**
- Resolve the dependency on-demand
- Subsequent calls return cached instance (for Singleton/Scoped)

**`di.GetProviderNamed[T](name string) container.Provider[T]`**
- Get a provider for a named dependency

### Runtime Registration API

**`di.RegisterRuntime(constructor interface{}, scope container.Scope) error`**
- Register a constructor after initial setup
- Automatically runs `di.Validate()`

**`di.RegisterRuntimeBatch(constructors []interface{}, scope container.Scope) error`**
- Batch runtime registration
- Runs `di.Validate()` once after all registrations

**`di.RegisterRuntimeWithScopes(registrations []ScopeRegistration) error`**
- Individual scoping for batch runtime registration
- Runs `di.Validate()` once

- Each constructor can have different lifetime
- Returns error if any registration fails

**`di.Reset()`**
- Clear all registrations and cached instances
- Primarily for testing purposes

Constructor requirements

- Must be a function.
- Return signature must be either:
  - T
  - (T, error) — when non-nil error is returned, resolution fails and the error is propagated.
- Parameters (if any) must be types that are also constructible by registered constructors.

## Dependency Scopes

### Singleton (Default)
- Created once and cached globally
- Same instance returned for all resolutions
- Best for: Configuration, database connections, services

### Transient
- New instance created on every resolution
- No caching
- Best for: Stateful operations, per-operation resources

### Scoped
- Created once per scope context
- Cached within the scope
- Different instances across different scopes
- Best for: Per-request data, per-session state

## Advanced Features

### Lazy Loading with Providers

Defer dependency creation until actually needed:

```go
type Service struct {
    dbProvider container.Provider[*Database]
}

func NewService(dbProvider container.Provider[*Database]) *Service {
    return &Service{dbProvider: dbProvider}
}

func (s *Service) DoWork() error {
    db, err := s.dbProvider.Get()  // Created only when needed
    if err != nil {
        return err
    }
    // Use db...
}
```

### Runtime Registration

Register dependencies dynamically after initialization:

```go
// Initialize base dependencies
di.Init([]interface{}{NewConfig, NewLogger})

// Later, register a single plugin
pluginConstructor := func(cfg *Config) *Plugin {
    return &Plugin{Config: cfg}
}

err := di.RegisterRuntime(pluginConstructor, container.Singleton)
plugin, _ := di.Resolve[*Plugin]()
```

**Batch Registration:**

```go
// Register multiple constructors with same scope
constructors := []interface{}{
    NewPluginA,
    NewPluginB,
    NewPluginC,
}

err := di.RegisterRuntimeBatch(constructors, container.Singleton)

// Or register with different scopes
registrations := []di.ScopeRegistration{
    {Constructor: NewPluginA, Scope: container.Singleton},
    {Constructor: NewPluginB, Scope: container.Transient},
    {Constructor: NewPluginC, Scope: container.Scoped},
}

err := di.RegisterRuntimeWithScopes(registrations)
```

### Interface-Based Resolution

Bind an interface to a concrete implementation:

```go
type Repository interface {
    Save(data string) error
}

type SQLRepo struct{}
func (r *SQLRepo) Save(data string) error { return nil }
func NewSQLRepo() *SQLRepo { return &SQLRepo{} }

// 1. Register concrete constructor
di.Init([]interface{}{NewSQLRepo})

// 2. Bind interface to concrete type
di.BindInterface[Repository, *SQLRepo]()

// 3. Resolve by interface
repo, _ := di.Resolve[Repository]()
```

**Multiple Implementations (Named Bindings):**

```go
di.BindInterfaceNamed[Logger, *ConsoleLogger]("console")
di.BindInterfaceNamed[Logger, *FileLogger]("file")

consoleLogger, _ := di.ResolveNamed[Logger]("console")
fileLogger, _ := di.ResolveNamed[Logger]("file")
```

### Eager Graph Validation

Verify your wiring at startup:

```go
func main() {
    di.Init(constructors)
    
    // Catch cycles/missing types before they happen at runtime
    if err := di.Validate(); err != nil {
        log.Fatalf("DI Error: %v", err)
    }
}
```

### Test Overrides & Mocking

Swap implementations without resetting the whole container:

```go
func TestService(t *testing.T) {
    di.Reset()
    di.Init([]interface{}{NewRealRepo})
    
    // Override with a mock for this test
    di.Override(NewMockRepo, container.Singleton)
    
    svc, _ := di.Resolve[*Service]()
    // svc now uses MockRepo
}
```

### HTTP Request Scoping Example

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Create scope for this request
    scopeID := di.CreateScope()
    defer di.DestroyScope(scopeID)
    
    // Resolve request-scoped dependencies
    ctx, _ := di.ResolveScoped[*RequestContext](scopeID)
    svc, _ := di.ResolveScoped[*RequestService](scopeID)
    
    // Use dependencies...
}
```

## Behavior

- **Order independent**: Register constructors in any order
- **Circular dependency detection**: Clear error messages with dependency chains
- **Thread-safe**: Concurrent registration and resolution
- **Singleton caching**: Double-checked locking for performance
- **Scope isolation**: Scoped instances are isolated per context

Examples

See example/example_test.go for a runnable example:

```
go test ./example -run Test
```

Error handling notes

- If no constructor is registered for a requested type, Resolve returns an error.
- If a constructor returns (T, error) and the error is non-nil, Resolve returns that error.
- If type casting fails internally (shouldn’t under normal use), Resolve returns an error.

## Architecture

For detailed architecture documentation including flowcharts and internal design, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Limitations

- Generic resolution requires specifying the exact type parameter, e.g., `di.Resolve[*MyType]()`.
- Scoped dependencies require explicit scope management - always call `DestroyScope()` to prevent memory leaks.

## Development

Run tests:

```
go test ./...
```

License

MIT
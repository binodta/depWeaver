# DepWeaver Architecture

## System Overview

DepWeaver is a dependency injection container that manages object lifecycles and resolves dependencies at runtime using reflection.

```mermaid
graph TB
    subgraph "Public API Layer"
        Init[di.Init]
        InitScope[di.InitWithScope]
        Resolve[di.Resolve]
        ResolveScoped[di.ResolveScoped]
        GetProvider[di.GetProvider]
        RegisterRuntime[di.RegisterRuntime]
        BindInterface[di.BindInterface]
        ResolveNamed[di.ResolveNamed]
        Validate[di.Validate]
        Override[di.Override]
    end
    
    subgraph "Container Layer"
        DC[DependencyContainer]
        RegMgr[Registration Manager]
        ResMgr[Resolution Manager]
        ScopeMgr[Scope Manager]
    end
    
    subgraph "Storage"
        Constructors[Constructors Map]
        Singletons[Singleton Cache]
        ScopedCache[Scoped Cache]
        Creating[Creating Tracker]
        Bindings[Interface Bindings]
        NamedCaches[Named Instance Caches]
    end
    
    Init --> RegMgr
    InitScope --> RegMgr
    Resolve --> ResMgr
    ResolveScoped --> ResMgr
    RegisterRuntime --> RegMgr
    GetProvider --> ResMgr
    Validate --> DC
    Override --> RegMgr
    
    RegMgr --> Constructors
    ResMgr --> Singletons
    ResMgr --> ScopedCache
    ResMgr --> Creating
    ResMgr --> Bindings
    ResMgr --> NamedCaches
    ScopeMgr --> ScopedCache
    ScopeMgr --> NamedCaches
    
    DC --> RegMgr
    DC --> ResMgr
    DC --> ScopeMgr
```

## Core Components

### 1. DependencyContainer

Central component that manages all dependencies.

**Key Fields:**
- `constructors`: Map of type → Registration (constructor + scope)
- `dependencies`: Singleton instance cache
- `scopedInstances`: Scoped instance cache (by scope ID)
- `creating`: Circular dependency detection tracker
- `resolutionStack`: Dependency chain for error reporting
- `interfaceBindings`: Unnamed interface → concrete type mapping
- `namedInterfaceBindings`: Named interface bindings (name → interface → type)
- `namedConstructors`: Named concrete type constructors
- `namedDependencies`: Named singleton instance cache
- `namedScopedInstances`: Named scoped instance cache

**Thread Safety:**
- Uses `sync.RWMutex` for concurrent access
- Read locks for cache lookups
- Write locks for instance creation

### 2. Registration

Holds constructor metadata.

```go
type Registration struct {
    constructor func(*DependencyContainer, string) (interface{}, error)
    scope       Scope
    paramTypes  []reflect.Type // Metadata for validation
}
```

### 3. Scope Types

```go
const (
    Singleton Scope = iota  // Created once, cached globally
    Transient              // Created every time
    Scoped                 // Created once per scope context
)
```

## Resolution Flow

### Singleton Resolution

```mermaid
flowchart TD
    Start([Resolve Type T]) --> CheckCache{In Singleton<br/>Cache?}
    CheckCache -->|Yes| ReturnCached[Return Cached Instance]
    CheckCache -->|No| AcquireLock[Acquire Write Lock]
    AcquireLock --> DoubleCheck{Still Not<br/>Cached?}
    DoubleCheck -->|Cached| ReleaseLock1[Release Lock]
    ReleaseLock1 --> ReturnCached
    DoubleCheck -->|Not Cached| CheckCircular{Creating<br/>This Type?}
    CheckCircular -->|Yes| CircularError[Return Circular<br/>Dependency Error]
    CheckCircular -->|No| MarkCreating[Mark as Creating]
    MarkCreating --> ReleaseLock2[Release Lock]
    ReleaseLock2 --> ResolveDeps[Resolve Dependencies]
    ResolveDeps --> CallConstructor[Call Constructor]
    CallConstructor --> AcquireLock2[Acquire Write Lock]
    AcquireLock2 --> UnmarkCreating[Unmark Creating]
    UnmarkCreating --> CacheInstance[Cache Instance]
    CacheInstance --> ReleaseLock3[Release Lock]
    ReleaseLock3 --> ReturnInstance[Return Instance]
    
    ReturnCached --> End([Done])
    ReturnInstance --> End
    CircularError --> End
```

### Transient Resolution

```mermaid
flowchart TD
    Start([Resolve Type T]) --> CheckCircular{Creating<br/>This Type?}
    CheckCircular -->|Yes| CircularError[Return Circular<br/>Dependency Error]
    CheckCircular -->|No| MarkCreating[Mark as Creating]
    MarkCreating --> ResolveDeps[Resolve Dependencies]
    ResolveDeps --> CallConstructor[Call Constructor]
    CallConstructor --> UnmarkCreating[Unmark Creating]
    UnmarkCreating --> ReturnInstance[Return New Instance]
    
    ReturnInstance --> End([Done])
    CircularError --> End
```

### Scoped Resolution

```mermaid
flowchart TD
    Start([Resolve Type T<br/>with Scope ID]) --> ValidateScope{Scope ID<br/>Provided?}
    ValidateScope -->|No| ScopeError[Return Scope<br/>Required Error]
    ValidateScope -->|Yes| CheckCache{In Scope<br/>Cache?}
    CheckCache -->|Yes| ReturnCached[Return Cached Instance]
    CheckCache -->|No| AcquireLock[Acquire Write Lock]
    AcquireLock --> DoubleCheck{Still Not<br/>Cached?}
    DoubleCheck -->|Cached| ReleaseLock1[Release Lock]
    ReleaseLock1 --> ReturnCached
    DoubleCheck -->|Not Cached| EnsureScope[Ensure Scope Exists]
    EnsureScope --> CheckCircular{Creating<br/>This Type?}
    CheckCircular -->|Yes| CircularError[Return Circular<br/>Dependency Error]
    CheckCircular -->|No| MarkCreating[Mark as Creating]
    MarkCreating --> ReleaseLock2[Release Lock]
    ReleaseLock2 --> ResolveDeps[Resolve Dependencies<br/>in Same Scope]
    ResolveDeps --> CallConstructor[Call Constructor]
    CallConstructor --> AcquireLock2[Acquire Write Lock]
    AcquireLock2 --> UnmarkCreating[Unmark Creating]
    UnmarkCreating --> CacheInScope[Cache in Scope]
    CacheInScope --> ReleaseLock3[Release Lock]
    ReleaseLock3 --> ReturnInstance[Return Instance]
    
    ReturnCached --> End([Done])
    ReturnInstance --> End
    ScopeError --> End
    CircularError --> End
```

### Interface-Based Resolution

```mermaid
flowchart TD
    Start([Resolve Interface I]) --> CheckBinding{Binding<br/>Exists?}
    CheckBinding -->|No| ContinueNormal[Continue with<br/>Normal Resolution]
    CheckBinding -->|Yes| ResolveConcrete[Resolve Concrete<br/>Type C]
    ResolveConcrete --> UseScope[Apply Original Scope]
    UseScope --> End([Done])
    
    ContinueNormal --> End
```

### Eager Graph Validation

DepWeaver employs a **Validation on Mutation** strategy. Every function that modifies the dependency graph—including `Init`, `RegisterRuntime`, `Override`, and `BindInterface`—automatically triggers a full graph validation. This ensures that the container never enters an invalid state (cycles or missing dependencies).

```mermaid
flowchart TD
    Start([Validate Graph]) --> Loop[Iterate All Constructors]
    Loop --> Visit[Visit Type T]
    Visit --> CheckCycle{In Progress?}
    CheckCycle -->|Yes| CycleErr[Return Cycle Error]
    CheckCycle -->|No| MarkProgress[Mark In Progress]
    MarkProgress --> CheckDeps[Visit Dependency Types]
    CheckDeps --> UnmarkProgress[Unmark In Progress]
    UnmarkProgress --> MarkVisited[Mark Visited]
    MarkVisited --> CheckMissing{Missing Constructor?}
    CheckMissing -->|Yes| MissingErr[Return Missing Error]
    CheckMissing -->|No| Next[Next Registration]
    
    Next --> End([Done])
```

### Test Overrides

```mermaid
flowchart TD
    Start([Override Type T]) --> ReplaceReg[Replace Constructor]
    ReplaceReg --> ClearCache[Clear Cache for T]
    ClearCache --> ClearSingletons[Remove T from Singleton Cache]
    ClearSingletons --> ClearScoped[Remove T from Scoped Caches]
    ClearScoped --> End([Done])
```

## Provider Pattern

```mermaid
sequenceDiagram
    participant User
    participant Provider
    participant Container
    participant Constructor
    
    User->>Container: GetProvider[T](scopeID)
    Container->>Provider: Create Provider
    Container-->>User: Return Provider
    
    Note over User,Provider: Lazy - No instance created yet
    
    User->>Provider: Get()
    Provider->>Container: ResolveWithScope(T, scopeID)
    Container->>Constructor: Call constructor
    Constructor-->>Container: Return instance
    Container-->>Provider: Return instance
    Provider-->>User: Return instance
    
    Note over User,Provider: Subsequent calls return cached instance
    
    User->>Provider: Get() again
    Provider->>Container: ResolveWithScope(T, scopeID)
    Container-->>Provider: Return cached instance
    Provider-->>User: Return instance
```

## Scope Lifecycle

```mermaid
sequenceDiagram
    participant Handler as HTTP Handler
    participant DI as DI Container
    participant Scope as Scope Manager
    participant Cache as Scoped Cache
    
    Handler->>DI: CreateScope()
    DI->>Scope: Generate UUID
    Scope->>Cache: Create empty map
    Scope-->>DI: Return scope ID
    DI-->>Handler: Return scope ID
    
    Handler->>DI: ResolveScoped[T](scopeID)
    DI->>Cache: Check scope cache
    Cache-->>DI: Not found
    DI->>DI: Create instance
    DI->>Cache: Store in scope
    DI-->>Handler: Return instance
    
    Handler->>DI: ResolveScoped[T](scopeID)
    DI->>Cache: Check scope cache
    Cache-->>DI: Found
    DI-->>Handler: Return cached instance
    
    Handler->>DI: DestroyScope(scopeID)
    DI->>Cache: Delete scope map
    DI-->>Handler: Done
```

## Circular Dependency Detection

```mermaid
flowchart TD
    Start([Resolve A]) --> MarkA[Mark A as Creating]
    MarkA --> ResolveB[Resolve Dependency B]
    ResolveB --> CheckB{B Creating?}
    CheckB -->|No| MarkB[Mark B as Creating]
    MarkB --> ResolveA[Resolve Dependency A]
    ResolveA --> CheckA{A Creating?}
    CheckA -->|Yes| Error[Circular Dependency Error:<br/>A → B → A]
    CheckA -->|No| Continue[Continue...]
    CheckB -->|Yes| Error2[Circular Dependency Error]
    
    Error --> End([Return Error])
    Error2 --> End
```

## Thread Safety Strategy

### Read-Write Lock Pattern

```mermaid
flowchart LR
    subgraph "Fast Path - Read Lock"
        R1[Acquire RLock]
        R2[Check Cache]
        R3{Found?}
        R4[Release RLock]
        R5[Return Instance]
        
        R1 --> R2 --> R3
        R3 -->|Yes| R4 --> R5
    end
    
    subgraph "Slow Path - Write Lock"
        W1[Acquire Lock]
        W2[Double Check]
        W3{Still Missing?}
        W4[Mark Creating]
        W5[Release Lock]
        W6[Create Instance]
        W7[Acquire Lock]
        W8[Cache Instance]
        W9[Release Lock]
        
        R3 -->|No| W1
        W1 --> W2 --> W3
        W3 -->|Yes| W4 --> W5 --> W6 --> W7 --> W8 --> W9
        W3 -->|No| W9
    end
```

### Concurrent Resolution

Multiple goroutines can:
- ✅ Read cached singletons concurrently (RLock)
- ✅ Create different types concurrently (separate locks per type)
- ✅ Resolve transients concurrently (no caching)
- ✅ Use different scopes concurrently (separate scope caches)

Only one goroutine creates a singleton instance (double-checked locking).

## Performance Characteristics

| Operation | First Call | Subsequent Calls | Concurrency |
|-----------|-----------|------------------|-------------|
| Singleton | Slow (construction + lock) | Fast (read lock) | High |
| Transient | Slow (construction) | Slow (construction) | Very High |
| Scoped | Slow (construction + lock) | Fast (read lock) | High (per scope) |
| Provider.Get() | Deferred to Get() | Same as scope | Same as scope |

## Memory Management

### Singleton Cache
- Lives for application lifetime
- Cleared only on `Reset()`

### Scoped Cache
- Lives for scope lifetime
- Cleared on `DestroyScope(scopeID)`
- Important: Always call `DestroyScope()` to prevent memory leaks

### Transient
- No caching
- Garbage collected when no references remain

## Error Handling

### Error Types

1. **No Constructor Registered**
   ```
   no constructor registered for type *MyType
   ```

2. **Circular Dependency**
   ```
   circular dependency detected: *ServiceA → *ServiceB → *ServiceA
   ```

3. **Scope Required**
   ```
   scope ID required for scoped dependency *RequestContext
   ```

4. **Constructor Error**
   ```
   error resolving dependency for *MyType: <original error>
   ```

All errors are propagated up the dependency chain with context.

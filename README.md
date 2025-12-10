DepWeaver

Lightweight, reflection-powered dependency injection for Go. Define constructors for your types and let DepWeaver build the dependency graph for you at runtime.

- Minimal API surface
- Constructor-based injection (order independent)
- Supports constructors that return either T or (T, error)
- Detects circular dependencies
- Thread-safe resolution and registration

Installation

Use the Go tool:

```
go get github.com/binodta/depWeaver
```

Quick start

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
    constructors := []interface{}{
        NewService,
        NewConfig,
        NewDB,          // (T, error) supported
        NewRepo,
    }

    di.Init(constructors)

    svc, err := di.Resolve[*Service]()
    if err != nil {
        panic(err)
    }
    fmt.Println("Service ready with DSN:", svc.Repo.DB.DSN)
}
```

API overview

- di.Init(constructors []interface{})
  - Register one or more constructor functions. Each constructor must be a function that returns either T or (T, error). Its parameters are other types managed by the container.

- di.Resolve[T]() (T, error)
  - Resolve and return an instance of T. Dependencies required by T’s constructor are resolved automatically.

Constructor requirements

- Must be a function.
- Return signature must be either:
  - T
  - (T, error) — when non-nil error is returned, resolution fails and the error is propagated.
- Parameters (if any) must be types that are also constructible by registered constructors.

Features and behavior

- Order independent: You can register constructors in any order.
- Circular dependency detection: The container detects cycles and returns an error.
- Thread-safety: Locks protect registration and resolution.
- Singleton semantics: Each type is constructed once and cached for subsequent resolutions.

Examples

See example/example_test.go for a runnable example:

```
go test ./example -run Test
```

Error handling notes

- If no constructor is registered for a requested type, Resolve returns an error.
- If a constructor returns (T, error) and the error is non-nil, Resolve returns that error.
- If type casting fails internally (shouldn’t under normal use), Resolve returns an error.

Limitations

- DepWeaver resolves by concrete types as returned by constructors. If you need interface-based resolution, return the interface type from your constructor (e.g., func NewRepo(db *DB) Repository).
- Generic resolution requires specifying the exact type parameter, e.g., di.Resolve[*MyType]().

Development

Run tests:

```
go test ./...
```

License

MIT
# DepWeaver

depWeaver is a lightweight and flexible dependency injection library for Go (Golang) projects. It is designed to
simplify the process of managing dependencies in your applications by allowing you to define and resolve dependencies at
runtime based on the functions' requirements.

### Installation

To install depWeaver, use the Go tool:

```shell
go get github.com/binodta/depWeaver

```

### Usage

Basic usage of depWeaver is as follows:

```
	constructors := []interface{}{
		NewUserService,
		NewConfig,
		NewDatabaseConnection,
		NewLoggerService,
		NewUserRepository,
	}
	
	
	// Initialize the dependency injection container
	di.Init(constructors)
	
	// Resolve the dependencies
	service, error := di.Resolve[*UserService]()
```
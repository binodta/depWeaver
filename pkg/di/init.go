package di

import (
	"depweaver/internal/container"
	"log"
)

var dependencyContainer = container.New()

// Init Register all constructors
// @Param constructors []interface{} - list of constructors
func Init(constructors []interface{}) {

	for _, constructor := range constructors {
		if err := dependencyContainer.RegisterConstructor(constructor); err != nil {
			log.Fatalf("Failed to register constructor: %v", err)
		}
	}
}

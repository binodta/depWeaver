package main

import (
	"sync"
	"testing"
	"time"

	"github.com/binodta/depWeaver/internal/container"
	"github.com/binodta/depWeaver/pkg/di"
)

type ConcurDatabase struct {
	ID int
}

var concurDbCounter int
var concurMu sync.Mutex

func NewConcurDatabase() *ConcurDatabase {
	concurMu.Lock()
	concurDbCounter++
	concurMu.Unlock()
	// Simulate slow connection
	time.Sleep(50 * time.Millisecond)
	return &ConcurDatabase{ID: concurDbCounter}
}

type ConcurRepository struct {
	db *ConcurDatabase
}

func NewConcurRepository(db *ConcurDatabase) *ConcurRepository {
	return &ConcurRepository{db: db}
}

func TestConcurrentSingletonResolution(t *testing.T) {
	di.Reset()
	concurDbCounter = 0

	di.MustInit([]interface{}{
		NewConcurDatabase,
		NewConcurRepository,
	})

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([]*ConcurRepository, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			repo, err := di.Resolve[*ConcurRepository]()
			results[idx] = repo
			errors[idx] = err
		}(i)
	}

	wg.Wait()

	// Check for errors
	for i, err := range errors {
		if err != nil {
			t.Fatalf("Goroutine %d failed to resolve: %v", i, err)
		}
	}

	// Verify all goroutines got the same singleton instance
	if len(results) == 0 {
		t.Fatal("No results")
	}
	firstRepo := results[0]
	if firstRepo == nil {
		t.Fatal("First repo is nil")
	}

	for i := 1; i < numGoroutines; i++ {
		if results[i] != firstRepo {
			t.Errorf("Goroutine %d got different instance: %+v vs %+v", i, results[i], firstRepo)
		}
	}

	if concurDbCounter != 1 {
		t.Errorf("Database constructor called %d times, expected 1", concurDbCounter)
	}
}

// Global types for concurrency test to avoid local type closure issues and collisions
type ConcurCycleA struct{ B *ConcurCycleB }
type ConcurCycleB struct{ A *ConcurCycleA }
type ConcurValidC struct{ D *ConcurValidD }
type ConcurValidD struct{}

func NewConcurCycleA(b *ConcurCycleB) *ConcurCycleA { return &ConcurCycleA{B: b} }
func NewConcurCycleB(a *ConcurCycleA) *ConcurCycleB { return &ConcurCycleB{A: a} }
func NewConcurValidC(d *ConcurValidD) *ConcurValidC { return &ConcurValidC{D: d} }
func NewConcurValidD() *ConcurValidD                { return &ConcurValidD{} }

func TestConcurrentCircularDependencyDetection(t *testing.T) {
	di.Reset()

	// Register valid C, D
	di.MustInit([]interface{}{NewConcurValidC, NewConcurValidD})

	var wg sync.WaitGroup
	wg.Add(2)

	var errA, errC error

	go func() {
		defer wg.Done()
		// Register A, B at runtime.
		di.RegisterRuntime(NewConcurCycleA, container.Singleton)
		// Registering B will trigger validation and detect the cycle with A
		errA = di.RegisterRuntime(NewConcurCycleB, container.Singleton)
	}()

	go func() {
		defer wg.Done()
		// Resolve valid C simultaneously
		time.Sleep(10 * time.Millisecond)
		_, errC = di.Resolve[*ConcurValidC]()
	}()

	wg.Wait()

	if errA == nil {
		t.Error("Expected circular dependency error for A/B, but got nil")
	}
	if errC != nil {
		t.Errorf("Expected no error for C, but got: %v", errC)
	}
}

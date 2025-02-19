package main

import (
	"depweaver/pkg/di"
	"fmt"
	"testing"
)

// Config Configuration management
type Config struct {
	DatabaseURL    string
	LogLevel       string
	MaxConnections int
}

// DatabaseConnection Mock database connection
type DatabaseConnection struct {
	Url           string
	MaxConnection int
}

// LoggerService Logging service
type LoggerService struct {
	Level string
}

// User repository with dependencies
type UserRepository struct {
	db     *DatabaseConnection
	logger *LoggerService
}

// User service using repository
type UserService struct {
	repo *UserRepository
}

// Helper constructors
func NewConfig() *Config {
	return &Config{
		DatabaseURL:    "postgres://user:password@localhost/mydb",
		LogLevel:       "INFO",
		MaxConnections: 10,
	}
}

func NewDatabaseConnection(config *Config) *DatabaseConnection {
	return &DatabaseConnection{
		Url:           config.DatabaseURL,
		MaxConnection: config.MaxConnections,
	}
}

func NewLoggerService(config *Config) *LoggerService {
	return &LoggerService{
		Level: config.LogLevel,
	}
}

func NewUserRepository(db *DatabaseConnection, logger *LoggerService) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func main() {
	// Register constructors (order doesn't matter)
	constructors := []interface{}{
		NewUserService,
		NewConfig,
		NewDatabaseConnection,
		NewLoggerService,
		NewUserRepository,
	}
	di.Init(constructors)
	service, _ := di.Resolve[*UserService]()
	fmt.Printf("UserService created successfully\n")
	fmt.Printf("Database URL: %s\n", service.repo.db.Url)
	fmt.Printf("Log Level: %s\n", service.repo.logger.Level)
}

func Test(t *testing.T) {
	main()
}

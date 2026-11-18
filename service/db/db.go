// Package db executes SQL and reads schema for database assets.
package db

import (
	"nextopsx/service/executor"
)

// Service exposes database queries, writes and schema reads.
type Service struct {
	resolver *executor.Resolver
}

// New builds the database service. resolver may be nil in tests that only
// exercise registry-backed drivers.
func New(resolver *executor.Resolver) *Service {
	return &Service{resolver: resolver}
}

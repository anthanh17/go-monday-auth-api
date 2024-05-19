package db

import "database/sql"

// Store provides all functions to executre db queries as transaction(future)
type Store struct {
	// db *sql.DB // impl transaction - future
	*Queries
}

// NewStore creates a new store
func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
	}
}

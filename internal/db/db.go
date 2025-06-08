package db

import (
	"database/sql"
	"fmt"
)

// Querier wraps a *sql.DB connection and provides helper methods for common query patterns.
type Querier struct {
	db *sql.DB
}

// NewQuerier creates a new Querier instance.
func NewQuerier(db *sql.DB) *Querier {
	return &Querier{db: db}
}

// QueryRow executes a query that is expected to return a single row.
// It takes the query string, a function to scan the row, and optional arguments.
func (q *Querier) QueryRow(query string, scanFunc func(*sql.Row) error, args ...interface{}) error {
	row := q.db.QueryRow(query, args...)
	if err := scanFunc(row); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}
	return nil
}

// QueryRows executes a query that is expected to return multiple rows.
// It takes the query string, a function to scan each row, and optional arguments.
// The scanFunc will be called for each row returned by the query.
func (q *Querier) QueryRows(query string, scanFunc func(*sql.Rows) error, args ...interface{}) error {
	rows, err := q.db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := scanFunc(rows); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}
	return nil
}

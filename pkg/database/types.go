package database

// Row represents a single record in the virtual table.
// It wraps the underlying data (likely a map[string]interface{}).
type Row interface {
	// Get returns the value of a field.
	// Supports dot notation for nested fields.
	Get(field string) (interface{}, error)
	// Primitive returns the underlying data structure.
	Primitive() interface{}
}

// RowIterator allows iterating over rows in a table.
type RowIterator interface {
	// Next advances the iterator. Returns false if no more rows or error.
	Next() bool
	// Row returns the current row.
	Row() Row
	// Error returns any error that occurred during iteration.
	Error() error
	// Close releases resources.
	Close() error
}

// Table represents a dataset that can be scanned.
type Table interface {
	// Iterate returns a new iterator for scanning the table.
	Iterate() (RowIterator, error)
}

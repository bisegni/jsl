package database

import (
	"fmt"
	"sync"
)

// Catalog manages a collection of named tables (databases)
type Catalog struct {
	tables map[string]Table
	mu     sync.RWMutex
}

// NewCatalog creates a new empty catalog
func NewCatalog() *Catalog {
	return &Catalog{
		tables: make(map[string]Table),
	}
}

// RegisterTable adds a table to the catalog
func (c *Catalog) RegisterTable(name string, t Table) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tables[name] = t
}

// GetTable retrieves a table by name
func (c *Catalog) GetTable(name string) (Table, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.tables[name]
	if !ok {
		return nil, fmt.Errorf("table '%s' not found", name)
	}
	return t, nil
}

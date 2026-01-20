package engine_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/engine"
	"github.com/bisegni/jsl/pkg/planner"
	"github.com/bisegni/jsl/pkg/query"
)

func runQuery(t *testing.T, table database.Table, sql string) []map[string]interface{} {
	t.Helper()
	q, err := query.ParseQuery(sql)
	if err != nil {
		t.Fatalf("Failed to parse query %q: %v", sql, err)
	}

	rootNode, err := planner.CreatePlan(q, table)
	if err != nil {
		t.Fatalf("Failed to create plan for %q: %v", sql, err)
	}

	executor := engine.NewExecutor()
	var buf bytes.Buffer
	if err := executor.Execute(rootNode, &buf); err != nil {
		t.Fatalf("Failed to execute query %q: %v", sql, err)
	}

	output := buf.String()
	if output == "" {
		return nil
	}

	var results []map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(output))
	for decoder.More() {
		var m map[string]interface{}
		if err := decoder.Decode(&m); err != nil {
			t.Fatalf("Failed to decode output of %q: %v", sql, err)
		}
		results = append(results, m)
	}
	return results
}

func TestQueryFunctionality(t *testing.T) {
	inventoryPath := "../../examples/inventory.json"
	table := database.NewJSONTable(inventoryPath)

	t.Run("Basic Projection", func(t *testing.T) {
		results := runQuery(t, table, "SELECT name, price WHERE id = 1")
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		if results[0]["name"] != "Laptop" {
			t.Errorf("Expected Laptop, got %v", results[0]["name"])
		}
		if results[0]["price"] != 1200.50 {
			t.Errorf("Expected 1200.50, got %v", results[0]["price"])
		}
	})

	t.Run("Aliases", func(t *testing.T) {
		results := runQuery(t, table, "SELECT name AS product_name, stock AS qty WHERE id = 6")
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		if results[0]["product_name"] != "Standing Desk" {
			t.Errorf("Expected Standing Desk, got %v", results[0]["product_name"])
		}
		if results[0]["qty"].(float64) != 5 {
			t.Errorf("Expected 5, got %v", results[0]["qty"])
		}
	})

	t.Run("Selection with AND", func(t *testing.T) {
		results := runQuery(t, table, "SELECT name WHERE category = 'Electronics' AND price < 1000")
		if len(results) != 2 { // Smartphone (800) and Monitor (300)
			t.Fatalf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("Selection with OR", func(t *testing.T) {
		results := runQuery(t, table, "SELECT name WHERE category = 'Furniture' OR category = 'Appliances'")
		if len(results) != 3 { // Desk Chair, standing Desk, Coffee Maker
			t.Fatalf("Expected 3 results, got %d", len(results))
		}
	})

	t.Run("Nested Object Access", func(t *testing.T) {
		results := runQuery(t, table, "SELECT name, supplier.country WHERE id = 4")
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		if results[0]["supplier.country"] != "Sweden" {
			t.Errorf("Expected Sweden, got %v", results[0]["supplier.country"])
		}
	})

	t.Run("Aggregations: Global COUNT and SUM", func(t *testing.T) {
		results := runQuery(t, table, "SELECT COUNT(name), SUM(stock)")
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		// Expect COUNT_name=8, SUM_stock=10+25+100+15+40+5+0+0 = 195
		if results[0]["COUNT_name"].(float64) != 8 {
			t.Errorf("Expected count 8, got %v", results[0]["COUNT_name"])
		}
		if results[0]["SUM_stock"].(float64) != 195 {
			t.Errorf("Expected sum 195, got %v", results[0]["SUM_stock"])
		}
	})

	t.Run("Aggregations: Global MIN, MAX, AVG", func(t *testing.T) {
		results := runQuery(t, table, "SELECT MIN(price), MAX(price), AVG(price)")
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		if results[0]["MIN_price"].(float64) != 0 {
			t.Errorf("Expected min 0, got %v", results[0]["MIN_price"])
		}
		if results[0]["MAX_price"].(float64) != 1200.50 {
			t.Errorf("Expected max 1200.50, got %v", results[0]["MAX_price"])
		}
		// Avg: (1200.5 + 800 + 50.99 + 150 + 300 + 450 + 0) / 7 = 2951.49 / 7 = 421.641...
		if results[0]["AVG_price"].(float64) < 421.64 || results[0]["AVG_price"].(float64) > 421.65 {
			t.Errorf("Expected avg approx 421.64, got %v", results[0]["AVG_price"])
		}
	})

	t.Run("Grouping: GROUP BY category", func(t *testing.T) {
		results := runQuery(t, table, "SELECT category, SUM(stock) GROUP BY category")
		if len(results) != 4 { // Electronics, Appliances, Furniture, Misc
			t.Fatalf("Expected 4 groups, got %d", len(results))
		}

		groupMap := make(map[string]float64)
		for _, r := range results {
			groupMap[r["category"].(string)] = r["SUM_stock"].(float64)
		}

		if groupMap["Electronics"] != 75 { // 10+25+40
			t.Errorf("Expected Electronics stock 75, got %f", groupMap["Electronics"])
		}
		if groupMap["Furniture"] != 20 { // 15+5
			t.Errorf("Expected Furniture stock 20, got %f", groupMap["Furniture"])
		}
		if groupMap["Appliances"] != 100 {
			t.Errorf("Expected Appliances stock 100, got %f", groupMap["Appliances"])
		}
	})

	t.Run("Implicit Array Path", func(t *testing.T) {
		// Test if we can filter by a tag in the array
		results := runQuery(t, table, "SELECT name WHERE tags = 'mobile'")
		if len(results) != 1 {
			t.Fatalf("Expected 1 result (Smartphone), got %d", len(results))
		}
		if results[0]["name"] != "Smartphone" {
			t.Errorf("Expected Smartphone, got %v", results[0]["name"])
		}
	})

	t.Run("Comparison Operators", func(t *testing.T) {
		t.Run("Inequality", func(t *testing.T) {
			results := runQuery(t, table, "SELECT name WHERE category = 'Furniture' AND stock != 5")
			if len(results) != 1 || results[0]["name"] != "Desk Chair" {
				t.Errorf("Expected Desk Chair, got %v", results)
			}
		})
		t.Run("Greater Equal", func(t *testing.T) {
			results := runQuery(t, table, "SELECT name WHERE price >= 800")
			if len(results) != 2 { // Laptop (1200.5), Smartphone (800)
				t.Errorf("Expected 2 results, got %d", len(results))
			}
		})
		t.Run("Contains", func(t *testing.T) {
			results := runQuery(t, table, "SELECT name WHERE tags ~= 'work'")
			if len(results) != 2 { // Laptop, Monitor
				t.Errorf("Expected 2 results, got %d", len(results))
			}
		})
		t.Run("CONTAINS Keyword", func(t *testing.T) {
			results := runQuery(t, table, "SELECT name WHERE tags CONTAINS 'home'")
			if len(results) != 1 || results[0]["name"] != "Coffee Maker" {
				t.Errorf("Expected Coffee Maker, got %v", results)
			}
		})
	})

	t.Run("Boolean Literals", func(t *testing.T) {
		results := runQuery(t, table, "SELECT name WHERE active = TRUE")
		if len(results) != 1 || results[0]["name"] != "Mystery Box" {
			t.Errorf("Expected Mystery Box, got %v", results)
		}
	})

	t.Run("Null and Missing Fields", func(t *testing.T) {
		// Aggregation over null/missing fields should skip them
		results := runQuery(t, table, "SELECT AVG(price) WHERE category = 'Misc'")
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		// Mystery Box has no price, Old Cable has price 0.0
		// AVG should be 0.0 (sum=0, count=1)
		if results[0]["AVG_price"].(float64) != 0.0 {
			t.Errorf("Expected AVG 0.0, got %v", results[0]["AVG_price"])
		}
	})

	t.Run("Subqueries", func(t *testing.T) {
		// Simple subquery to unroll/rename
		results := runQuery(t, table, "SELECT p FROM (SELECT price AS p FROM table WHERE category='Furniture') WHERE p > 200")
		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}
		if results[0]["p"].(float64) != 450.0 {
			t.Errorf("Expected 450.0, got %v", results[0]["p"])
		}
	})

	t.Run("Empty Results", func(t *testing.T) {
		results := runQuery(t, table, "SELECT name WHERE id = 999")
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})
}

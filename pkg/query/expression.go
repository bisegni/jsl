package query

import (
	"strings"

	"github.com/bisegni/jsl/pkg/parser"
)

// Expression is a boolean expression that can be evaluated against a record
type Expression interface {
	Evaluate(record parser.Record) bool
}

// Condition is a simple filter (leaf node)
type Condition struct {
	Filter *Filter
}

func (c *Condition) Evaluate(record parser.Record) bool {
	return c.Filter.Match(record)
}

// AndExpression represents Logical AND
type AndExpression struct {
	Left  Expression
	Right Expression
}

func (a *AndExpression) Evaluate(record parser.Record) bool {
	return a.Left.Evaluate(record) && a.Right.Evaluate(record)
}

// OrExpression represents Logical OR
type OrExpression struct {
	Left  Expression
	Right Expression
}

func (o *OrExpression) Evaluate(record parser.Record) bool {
	return o.Left.Evaluate(record) || o.Right.Evaluate(record)
}

// ParseExpression parses a boolean expression string (e.g., "A=1 AND B=2")
// Precedence: AND binds tighter than OR?
// SQL precedence: NOT > AND > OR.
// Simple recursive descent or split strategy.
// Strategy: Split by " OR " first.
func ParseExpression(input string) Expression {
	input = strings.TrimSpace(input)

	// 1. Split by OR (lowest precedence)
	// We need to be careful not to split inside quotes.
	// For simplicity, assuming operators are surrounded by spaces or distinct.
	// Use case-insensitive splitter.
	orParts := splitByOperator(input, " OR ")
	if len(orParts) > 1 {
		expr := ParseExpression(orParts[0])
		for i := 1; i < len(orParts); i++ {
			expr = &OrExpression{
				Left:  expr,
				Right: ParseExpression(orParts[i]),
			}
		}
		return expr
	}

	// 2. Split by AND (higher precedence)
	andParts := splitByOperator(input, " AND ")
	if len(andParts) > 1 {
		expr := ParseExpression(andParts[0])
		for i := 1; i < len(andParts); i++ {
			expr = &AndExpression{
				Left:  expr,
				Right: ParseExpression(andParts[i]),
			}
		}
		return expr
	}

	// 3. Leaf node (Filter)
	// If it's wrapped in parens, unwrap and parse recursively
	if strings.HasPrefix(input, "(") && strings.HasSuffix(input, ")") {
		return ParseExpression(input[1 : len(input)-1])
	}

	filterExpr := ParseFilterExpression(input)
	if filterExpr == nil {
		// Fallback or error? For now, return a False condition or panic?
		// We'll return a Condition that always fails if invalid, or handle error.
		// Let's rely on ParseFilterExpression returning nil and existing logic handling it?
		// Currently returning valid objects.
		// If nil, maybe just return a dummy false condition.
		return &Condition{
			Filter: &Filter{Field: "error", Operator: "=", Value: "invalid"},
		}
	}
	return &Condition{
		Filter: NewFilter(filterExpr.Field, filterExpr.Operator, filterExpr.Value),
	}
}

// splitByOperator splits string by operator, ignoring quotes context if possible
// For this iteration, simple Case Insensitive Split is used.
func splitByOperator(s, op string) []string {
	// Normalized split (hacky but works for standard spacing)
	// Limitation: doesn't handle "field=' OR '", but that's a known limitation of simple splitting
	// Need a proper tokenizer for specific syntax robustness.
	// Given typical usage: "field=val OR field2=val2"

	// Case insensitive split
	upper := strings.ToUpper(s)
	upperOp := strings.ToUpper(op)

	parts := strings.Split(upper, upperOp)

	if len(parts) == 1 {
		return []string{s}
	}

	result := make([]string, 0, len(parts))
	lastPos := 0
	for _, p := range parts {
		// Reconstruct original casing
		// length of part p matches length of segment in s
		// Logic:
		// original part is s[lastPos : lastPos+len(p)]
		// new lastPos is lastPos + len(p) + len(op)

		segment := s[lastPos : lastPos+len(p)]
		result = append(result, strings.TrimSpace(segment))
		lastPos += len(p) + len(op)
	}

	return result
}

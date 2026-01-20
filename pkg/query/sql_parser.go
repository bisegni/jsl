package query

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Field represents a selected field with optional alias and aggregation
type Field struct {
	Path      string
	Alias     string
	Aggregate string // "MAX", "MIN", "AVG", "COUNT", "SUM" or empty
}

func (f Field) String() string {
	s := f.Path
	if f.Aggregate != "" {
		s = fmt.Sprintf("%s(%s)", f.Aggregate, f.Path)
	}
	if f.Alias != "" && f.Alias != f.Path {
		s += " AS " + f.Alias
	}
	return s
}

// SelectQuery represents a parsed SQL-like query IR (Intermediate Representation)
type SelectQuery struct {
	Fields    []Field
	FromTable string       // Name of the table if source is a table
	FromQuery *SelectQuery // Recursive subquery if source is another query
	Filter    Expression   // Compiled expression tree for the WHERE clause
	GroupBy   string
}

// Lexer definition
var (
	sqlLexer = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Keyword", Pattern: `(?i)\b(SELECT|FROM|WHERE|GROUP|BY|AS|AND|OR|TRUE|FALSE|CONTAINS)\b`},
		{Name: "Ident", Pattern: `[a-zA-Z_][a-zA-Z0-9_]*`},
		{Name: "Number", Pattern: `[-+]?\d*\.?\d+`},
		{Name: "String", Pattern: `'[^']*'|"[^"]*"`},
		{Name: "Operator", Pattern: `>=|<=|!=|~=|\.\.|[=<>!~]`},
		{Name: "Punct", Pattern: `[-+/*%,.()]`},
		{Name: "Whitespace", Pattern: `\s+`},
	})

	// Participle Parser
	sqlParser = participle.MustBuild[ASTSelect](
		participle.Lexer(sqlLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Keyword"),
		participle.Elide("Whitespace"),
		participle.UseLookahead(2), // Lookahead to resolve ambiguity if needed
	)
)

// ParseQuery parses a SELECT string using Participle
func ParseQuery(input string) (*SelectQuery, error) {
	// Pre-process? Participle handles whitespace.
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty query")
	}

	ast, err := sqlParser.ParseString("", input)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	return ast.ToSelectQuery(), nil
}

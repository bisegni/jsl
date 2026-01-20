package query

import (
	"fmt"
	"strings"
)

// AST for Participle Parser

type ASTSelect struct {
	SelectFields []*ASTSelectField `parser:"'SELECT' @@ (',' @@)*"`
	From         *ASTFromClause    `parser:"('FROM' @@)?"`
	Where        *ASTExpression    `parser:"('WHERE' @@)?"`
	GroupBy      *ASTValue         `parser:"('GROUP' 'BY' @@)?"`
}

type ASTSelectField struct {
	Expression *ASTExpression `parser:"@@"`
	Alias      string         `parser:"('AS' @Ident)?"`
}

type ASTFromClause struct {
	TableName *string    `parser:"(@Ident | @String)"`
	SubQuery  *ASTSelect `parser:"| '(' @@ ')'"`
}

type ASTExpression struct {
	Or []*ASTOrCondition `parser:"@@ ('OR' @@)*"`
}

type ASTOrCondition struct {
	And []*ASTCondition `parser:"@@ ('AND' @@)*"`
}

type ASTCondition struct {
	Grouped *ASTExpression      `parser:"  '(' @@ ')'"`
	Simple  *ASTSimpleCondition `parser:"| @@"`
}

type ASTSimpleCondition struct {
	Operand *ASTOperand `parser:"  @@"`
	Op      *string     `parser:"( @('='|'!='|'>'|'<'|'>='|'<='|'CONTAINS')"`
	Value   *ASTOperand `parser:"  @@ )?"`
}

type ASTOperand struct {
	Function *ASTFunction `parser:"  @@"`
	Literal  *ASTLiteral  `parser:"| @@"`
	Value    *ASTValue    `parser:"| @@"`
	SubQuery *ASTSelect   `parser:"| '(' @@ ')'"`
}

type ASTFunction struct {
	Name string        `parser:"@Ident"`
	Args []*ASTOperand `parser:"'(' @@ (',' @@)* ')'"`
}

type ASTValue struct {
	// Value can be a path with dots and wildcards
	// Ident or "*" separated by "."
	// We need to capture the whole thing as a string or list of parts?
	// Simplest: Capture parts and join them.
	Parts []string `parser:"(@Ident | @('*')) ('.' (@Ident | @('*')))*"`
}

func (v *ASTValue) String() string {
	return strings.Join(v.Parts, ".")
}

type ASTLiteral struct {
	Number *float64 `parser:"@Number"`
	StrVal *string  `parser:"| @String"`
	Bool   *bool    `parser:"| @('TRUE'|'FALSE')"`
}

// Helpers

func (s *ASTSelect) ToSelectQuery() *SelectQuery {
	sq := &SelectQuery{
		Fields: []Field{},
	}

	for _, f := range s.SelectFields {
		path, agg := f.Info()

		alias := f.Alias
		if alias == "" {
			if agg != "" {
				alias = fmtKey(agg, path)
			} else {
				alias = path
			}
		}

		sq.Fields = append(sq.Fields, Field{
			Path:      path,
			Alias:     alias,
			Aggregate: agg,
		})
	}

	if s.From != nil {
		if s.From.TableName != nil {
			sq.FromTable = *s.From.TableName
		} else if s.From.SubQuery != nil {
			sq.FromQuery = s.From.SubQuery.ToSelectQuery()
		}
	}

	if s.GroupBy != nil {
		sq.GroupBy = s.GroupBy.String()
	}

	if s.Where != nil {
		sq.Filter = s.Where.ToExpression()
	}

	return sq
}

func (f *ASTSelectField) Info() (path, agg string) {
	if f.Expression == nil {
		return "", ""
	}

	if len(f.Expression.Or) > 0 && len(f.Expression.Or[0].And) > 0 {
		cond := f.Expression.Or[0].And[0]
		if cond.Grouped != nil {
			// recursively call info? Select fields usually aren't grouped expressions,
			// they are paths or functions.
			// For now, only handle simple.
		} else if cond.Simple != nil && cond.Simple.Operand != nil {
			op := cond.Simple.Operand
			if op.Function != nil {
				agg = op.Function.Name
				if len(op.Function.Args) > 0 {
					path, _ = op.Function.Args[0].getSimplePath()
				}
			} else if op.Value != nil {
				path = op.Value.String()
			}
		}
	}
	return
}

func (o *ASTOperand) getSimplePath() (string, string) {
	if o.Value != nil {
		return o.Value.String(), ""
	}
	return "", ""
}

func (e *ASTExpression) String() string {
	var parts []string
	for _, or := range e.Or {
		parts = append(parts, or.String())
	}
	return strings.Join(parts, " OR ")
}

func (o *ASTOrCondition) String() string {
	var parts []string
	for _, and := range o.And {
		parts = append(parts, and.String())
	}
	return strings.Join(parts, " AND ")
}

func (c *ASTCondition) String() string {
	if c.Grouped != nil {
		return "(" + c.Grouped.String() + ")"
	}
	if c.Simple != nil {
		s := c.Simple.Operand.String()
		if c.Simple.Op != nil && c.Simple.Value != nil {
			s += *c.Simple.Op + c.Simple.Value.String()
		}
		return s
	}
	return ""
}

func (o *ASTOperand) String() string {
	if o.Function != nil {
		return o.Function.String()
	}
	if o.Literal != nil {
		return o.Literal.String()
	}
	if o.Value != nil {
		return o.Value.String()
	}
	if o.SubQuery != nil {
		// Reconstruct subquery? Not supported in simple expressions yet
		return "(SUBQUERY)"
	}
	return ""
}

func (f *ASTFunction) String() string {
	var args []string
	for _, a := range f.Args {
		args = append(args, a.String())
	}
	return fmt.Sprintf("%s(%s)", f.Name, strings.Join(args, ","))
}

func (l *ASTLiteral) String() string {
	if l.Number != nil {
		return fmt.Sprintf("%v", *l.Number)
	}
	if l.StrVal != nil {
		return fmt.Sprintf("'%s'", *l.StrVal) // simplistic quoting
	}
	if l.Bool != nil {
		return fmt.Sprintf("%v", *l.Bool)
	}
	return ""
}

func (l *ASTLiteral) ToValue() interface{} {
	if l.Number != nil {
		return *l.Number
	}
	if l.StrVal != nil {
		return *l.StrVal
	}
	if l.Bool != nil {
		return *l.Bool
	}
	return nil
}

func fmtKey(agg, path string) string {
	return agg + "_" + strings.ReplaceAll(path, ".", "_")
}

// Map AST to Expression interface

func (e *ASTExpression) ToExpression() Expression {
	if len(e.Or) == 0 {
		return nil
	}
	var expr Expression = e.Or[0].ToExpression()
	for i := 1; i < len(e.Or); i++ {
		expr = &OrExpression{
			Left:  expr,
			Right: e.Or[i].ToExpression(),
		}
	}
	return expr
}

func (o *ASTOrCondition) ToExpression() Expression {
	if len(o.And) == 0 {
		return nil
	}
	var expr Expression = o.And[0].ToExpression()
	for i := 1; i < len(o.And); i++ {
		expr = &AndExpression{
			Left:  expr,
			Right: o.And[i].ToExpression(),
		}
	}
	return expr
}

func (c *ASTCondition) ToExpression() Expression {
	if c.Grouped != nil {
		return c.Grouped.ToExpression()
	}
	if c.Simple != nil {
		// Map to Filter
		leftPath := c.Simple.Operand.String() // simplify
		op := "="
		if c.Simple.Op != nil {
			op = *c.Simple.Op
		}
		var val interface{}
		if c.Simple.Value != nil {
			val = c.Simple.Value.ToValue()
		}

		return &Condition{
			Filter: NewFilter(leftPath, op, val),
		}
	}
	return nil
}

func (o *ASTOperand) UnquotedString() string {
	if o.Literal != nil {
		if o.Literal.StrVal != nil {
			return *o.Literal.StrVal
		}
		return o.Literal.String()
	}
	return o.String()
}

func (o *ASTOperand) ToValue() interface{} {
	if o.Literal != nil {
		return o.Literal.ToValue()
	}
	if o.Value != nil {
		return o.Value.String()
	}
	// Functions and Subqueries not supported as filter values yet
	return o.String()
}

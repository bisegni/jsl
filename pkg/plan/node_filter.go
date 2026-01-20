package plan

import (
	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/query"
)

// FilterNode filters rows based on an expression
type FilterNode struct {
	Input      Node
	Expression query.Expression
}

func (n *FilterNode) Execute() (database.RowIterator, error) {
	inputIter, err := n.Input.Execute()
	if err != nil {
		return nil, err
	}
	return &filterIterator{source: inputIter, expression: n.Expression}, nil
}

func (n *FilterNode) Children() []Node {
	return []Node{n.Input}
}

func (n *FilterNode) Explain() string {
	return "Filter(expression: " + n.Expression.String() + ")"
}

package plan

import (
	"fmt"
	"strings"

	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/query"
)

// ProjectNode projects fields
type ProjectNode struct {
	Input  Node
	Fields []query.Field
	Filter query.Expression
}

func (n *ProjectNode) Execute() (database.RowIterator, error) {
	inputIter, err := n.Input.Execute()
	if err != nil {
		return nil, err
	}
	return &projectIterator{source: inputIter, fields: n.Fields, filter: n.Filter}, nil
}

func (n *ProjectNode) Children() []Node {
	return []Node{n.Input}
}

func (n *ProjectNode) Explain() string {
	var fieldStrings []string
	for _, f := range n.Fields {
		fieldStrings = append(fieldStrings, f.String())
	}
	return fmt.Sprintf("Project(%s)", strings.Join(fieldStrings, ", "))
}

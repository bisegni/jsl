package plan

import (
	"fmt"

	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/query"
)

// ProjectNode projects fields
type ProjectNode struct {
	Input  Node
	Fields []query.Field
}

func (n *ProjectNode) Execute() (database.RowIterator, error) {
	inputIter, err := n.Input.Execute()
	if err != nil {
		return nil, err
	}
	return &projectIterator{source: inputIter, fields: n.Fields}, nil
}

func (n *ProjectNode) Children() []Node {
	return []Node{n.Input}
}

func (n *ProjectNode) Explain() string {
	return fmt.Sprintf("Project(%d fields)", len(n.Fields))
}

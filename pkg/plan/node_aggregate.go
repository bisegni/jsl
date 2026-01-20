package plan

import (
	"fmt"

	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/query"
)

// AggregateNode handles GroupBy and Aggregations
type AggregateNode struct {
	Input        Node
	GroupByField string
	Fields       []query.Field
}

func (n *AggregateNode) Execute() (database.RowIterator, error) {
	// We need to implement the aggregation logic here or delegate to a separate implementation
	// For now, let's assume we implement `aggregateIterator` in this package.
	return &aggregateIterator{
		input:        n.Input,
		groupByField: n.GroupByField,
		fields:       n.Fields,
	}, nil
}

func (n *AggregateNode) Children() []Node {
	return []Node{n.Input}
}

func (n *AggregateNode) Explain() string {
	return fmt.Sprintf("Aggregate(Group: %s)", n.GroupByField)
}

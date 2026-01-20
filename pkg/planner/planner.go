package planner

import (
	"github.com/bisegni/jsl/pkg/database"
	"github.com/bisegni/jsl/pkg/plan"
	"github.com/bisegni/jsl/pkg/query"
)

// CreatePlan converts a Query IR into an Execution Plan
func CreatePlan(q *query.SelectQuery, rootTable database.Table) (plan.Node, error) {
	// 1. Resolve Input (FROM)
	var inputNode plan.Node

	if q.FromQuery != nil {
		// Recursive subquery
		subPlan, err := CreatePlan(q.FromQuery, rootTable)
		if err != nil {
			return nil, err
		}
		inputNode = subPlan
	} else if q.FromTable != "" {
		// Named table
		inputNode = &plan.ScanNode{TableName: q.FromTable, Table: rootTable}
	} else {
		// Default input
		inputNode = &plan.ScanNode{TableName: "default", Table: rootTable}
	}

	var currentNode plan.Node = inputNode

	// 2. Apply WHERE (Filter)
	if q.Filter != nil {
		currentNode = &plan.FilterNode{
			Input:      currentNode,
			Expression: q.Filter,
		}
	}

	// 3. Apply GroupBy / Aggregation
	hasAggregation := q.GroupBy != ""
	if !hasAggregation {
		for _, f := range q.Fields {
			if f.Aggregate != "" {
				hasAggregation = true
				break
			}
		}
	}

	if hasAggregation {
		currentNode = &plan.AggregateNode{
			Input:        currentNode,
			GroupByField: q.GroupBy,
			Fields:       q.Fields,
		}
	} else if len(q.Fields) > 0 {
		// Projection
		currentNode = &plan.ProjectNode{
			Input:  currentNode,
			Fields: q.Fields,
		}
	}

	return currentNode, nil
}

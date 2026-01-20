package plan

import (
	"github.com/bisegni/jsl/pkg/database"
)

// Node represents an execution node in the query plan
type Node interface {
	Execute() (database.RowIterator, error)
	Children() []Node
	Explain() string
}

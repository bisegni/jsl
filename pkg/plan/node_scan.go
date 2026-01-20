package plan

import (
	"fmt"

	"github.com/bisegni/jsl/pkg/database"
)

// ScanNode scans a table
type ScanNode struct {
	TableName string
	Table     database.Table
}

func (n *ScanNode) Execute() (database.RowIterator, error) {
	return n.Table.Iterate()
}

func (n *ScanNode) Children() []Node {
	return nil
}

func (n *ScanNode) Explain() string {
	return fmt.Sprintf("Scan(table: %s)", n.TableName)
}

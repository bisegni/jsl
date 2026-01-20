package engine

import (
	"encoding/json"
	"io"

	"github.com/bisegni/jsl/pkg/plan"
)

// Executor runs a Query Plan
type Executor struct {
	Pretty bool
}

func NewExecutor() *Executor {
	return &Executor{
		Pretty: false,
	}
}

// Execute runs the query plan and writes output
func (e *Executor) Execute(rootNode plan.Node, w io.Writer) error {
	// Execute the Plan
	iterator, err := rootNode.Execute()
	if err != nil {
		return err
	}
	defer iterator.Close()

	// Stream results
	encoder := json.NewEncoder(w)
	if e.Pretty {
		encoder.SetIndent("", "  ")
	} else {
		encoder.SetIndent("", "")
	}

	for iterator.Next() {
		row := iterator.Row().Primitive()
		if err := encoder.Encode(row); err != nil {
			return err
		}
	}

	if err := iterator.Error(); err != nil {
		return err
	}

	return nil
}

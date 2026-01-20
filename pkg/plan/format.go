package plan

import (
	"strings"
)

// FormatPlan generates a visual string representation of the plan tree
func FormatPlan(n Node) string {
	var sb strings.Builder
	formatRecursive(n, "", true, &sb)
	return sb.String()
}

func formatRecursive(n Node, prefix string, checkLast bool, sb *strings.Builder) {
	// Current node
	sb.WriteString(prefix)
	if checkLast {
		sb.WriteString("└─ ")
		prefix += "   "
	} else {
		sb.WriteString("├─ ")
		prefix += "│  "
	}
	sb.WriteString(n.Explain())
	sb.WriteString("\n")

	// Children
	children := n.Children()
	for i, child := range children {
		isLast := i == len(children)-1
		formatRecursive(child, prefix, isLast, sb)
	}
}

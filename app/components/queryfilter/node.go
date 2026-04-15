package queryfilter

// Node is the AST interface for RSQL filter expressions.
type Node interface {
	isNode()
}

// ComparisonNode represents a single field comparison: field op value.
type ComparisonNode struct {
	Field string
	Op    Op
	Value string
}

func (*ComparisonNode) isNode() {}

// LogicalNode represents a logical AND or OR combination of two sub-expressions.
type LogicalNode struct {
	Op    LogicalOp
	Left  Node
	Right Node
}

func (*LogicalNode) isNode() {}

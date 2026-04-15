package queryfilter

// Op is a comparison operator string.
type Op = string

const (
	OpEq     Op = "eq"
	OpNe     Op = "ne"
	OpGt     Op = "gt"
	OpLt     Op = "lt"
	OpGte    Op = "gte"
	OpLte    Op = "lte"
	OpPrefix Op = "prefix"
	OpSuffix Op = "suffix"
	OpLike   Op = "like"
)

// LogicalOp is a logical operator string.
type LogicalOp = string

const (
	LogicalAND LogicalOp = "AND"
	LogicalOR  LogicalOp = "OR"
)

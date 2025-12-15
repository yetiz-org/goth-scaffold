package models

type CassandraModelTTL interface {
	EntityTTL() int
}

type CassandraModelTimestamp interface {
	EntityTimestamp() int
}

type CassandraModelIf interface {
	EntityIf() map[string]any
}

type CassandraTTL struct {
	TTL int `json:"-"`
}

func (c *CassandraTTL) EntityTTL() int {
	return c.TTL
}

type CassandraTimestamp struct {
	Timestamp int `json:"-"`
}

func (c *CassandraTimestamp) EntityTimestamp() int {
	return c.Timestamp
}

type CassandraIf struct {
	If map[string]any `json:"-"`
}

func (c *CassandraIf) EntityIf() map[string]any {
	return c.If
}

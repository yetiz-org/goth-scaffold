package models

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/gocql/gocql"
	buf "github.com/yetiz-org/goth-bytebuf"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-util/value"
)

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

const CassandraQueryOptionFlag = "GOTH_QUERY_OPTION"

type CassandraQueryOption interface {
	Operate(query *gocql.Query)
}

func DefaultCassandraQueryOption(next string) []CassandraQueryOption {
	return []CassandraQueryOption{&CassandraQueryNext{Next: next}, &CassandraQueryLimit{Limit: 50}}
}

type CassandraQueryLimit struct {
	Limit int
}

func (q *CassandraQueryLimit) Operate(query *gocql.Query) {
	if q.Limit > 0 {
		if _, f := query.Context().Value(CassandraQueryOptionFlag).(map[string]any)["PageState"]; !f {
			query.PageState(nil)
		}

		query.PageSize(q.Limit)
	}
}

type CassandraQueryNext struct {
	Next string
}

func (q *CassandraQueryNext) Operate(query *gocql.Query) {
	if q.Next != "" {
		bs := buf.NewByteBuf(IDCodec.DecodeString(q.Next))
		l := bs.ReadInt32()
		query.Context().Value(CassandraQueryOptionFlag).(map[string]any)["PageState"] = true
		query.PageState(bs.ReadBytes(int(l)))
	}
}

func scanNext[T Model](iter *gocql.Iter) (instance T, ok bool) {
	instance = reflect.New(reflect.New(reflect.TypeOf(*new(T))).Elem().Type().Elem()).Interface().(T)
	if model, isCustomScan := any(instance).(CassandraModelScan); isCustomScan {
		ok = model.Scan(iter)
	} else {
		ok = scan(iter, instance)
	}

	return instance, ok
}

func scan(iter *gocql.Iter, m any) bool {
	maps := map[string]any{}
	if !iter.MapScan(maps) {
		return false
	}

	if err := json.Unmarshal([]byte(value.JsonMarshal(maps)), m); err != nil {
		kklogger.ErrorJ("models:Helper.scan#keyspaces!scan_error", err.Error())
		return false
	}

	return true
}

func prepareQuery(query *gocql.Query, opts ...CassandraQueryOption) *gocql.Query {
	query = query.WithContext(context.WithValue(context.Background(), CassandraQueryOptionFlag, map[string]any{}))
	for _, opt := range opts {
		opt.Operate(query)
	}

	return query
}

func CassandraQueryFinalizeFirst[T Model](session *gocql.Session, stmt string, args []any, opts ...CassandraQueryOption) (object T) {
	query := prepareQuery(session.Query(stmt, args...), append(opts, &CassandraQueryLimit{Limit: 1})...)
	iter := query.Iter()
	defer CassandraCloseIter(iter)

	if instance, ok := scanNext[T](iter); ok {
		object = instance
	}

	return
}

func CassandraQueryFinalize[T Model](session *gocql.Session, stmt string, args []any, opts ...CassandraQueryOption) (objects []T, result CassandraQueryResult[T]) {
	query := prepareQuery(session.Query(stmt, args...), opts...)
	objects = make([]T, 0)
	iter := query.Iter()
	defer CassandraCloseIter(iter)

	for instance, ok := scanNext[T](iter); ok; instance, ok = scanNext[T](iter) {
		objects = append(objects, instance)
	}

	if l := int32(len(iter.PageState())); l > 0 {
		result.NextId = IDCodec.EncodeToString(buf.EmptyByteBuf().WriteInt32(l).WriteBytes(iter.PageState()).Bytes())
	}

	result.Count = len(objects)
	result.session = session
	result.queryOptions = opts
	result.query = query
	return
}

type CassandraQueryResult[T Model] struct {
	Count        int    `json:"count"`
	NextId       string `json:"next_id"`
	session      *gocql.Session
	queryOptions []CassandraQueryOption
	query        *gocql.Query
}

func (q *CassandraQueryResult[T]) Next(session ...*gocql.Session) (objects []T, result CassandraQueryResult[T]) {
	if q.query == nil || q.NextId == "" {
		return
	}

	newQueryOptions := make([]CassandraQueryOption, 0)
	for _, opt := range q.queryOptions {
		if _, ok := opt.(*CassandraQueryNext); ok {
			continue
		}

		newQueryOptions = append(newQueryOptions, opt)
	}

	sess := q.session
	if len(session) > 0 {
		sess = session[0]
	}

	newQueryOptions = append(newQueryOptions, &CassandraQueryNext{Next: q.NextId})
	return CassandraQueryFinalize[T](sess, q.query.Statement(), q.query.Values(), newQueryOptions...)
}

func CassandraCloseIter(iter *gocql.Iter) {
	if iter != nil {
		if err := iter.Close(); err != nil {
			kklogger.WarnJ("models:Helper.CloseIter#keyspaces!close_error", err.Error())
		}
	}
}

type CassandraQueryBuilder[T Model] struct {
	session      *gocql.Session
	fields       []string
	conditions   map[string]any
	orders       []string
	queryOptions []CassandraQueryOption
	limit        int
	nextId       string
}

func NewCassandraQueryBuilder[T Model](session *gocql.Session) *CassandraQueryBuilder[T] {
	return &CassandraQueryBuilder[T]{session: session, conditions: map[string]any{}}
}

func (b *CassandraQueryBuilder[T]) Fields(fields ...string) *CassandraQueryBuilder[T] {
	b.fields = fields
	return b
}

func (b *CassandraQueryBuilder[T]) Where(condition string, arg any) *CassandraQueryBuilder[T] {
	b.conditions[condition] = arg
	return b
}

func (b *CassandraQueryBuilder[T]) Order(order string) *CassandraQueryBuilder[T] {
	b.orders = append(b.orders, order)
	return b
}

func (b *CassandraQueryBuilder[T]) Limit(limit int) *CassandraQueryBuilder[T] {
	b.limit = limit
	return b
}

func (b *CassandraQueryBuilder[T]) Next(nextId string) *CassandraQueryBuilder[T] {
	b.nextId = nextId
	return b
}

func (b *CassandraQueryBuilder[T]) buildQuery() (stmt string, args []any) {
	stmt = "SELECT "
	args = make([]any, 0)

	if len(b.fields) > 0 {
		stmt += strings.Join(b.fields, ",")
	} else {
		stmt += "*"
	}

	stmt += " FROM " + (reflect.New(reflect.New(reflect.TypeOf(*new(T))).Elem().Type().Elem()).Interface()).(T).TableName()

	if len(b.conditions) > 0 {
		conditions := make([]string, 0)
		for k, v := range b.conditions {
			conditions = append(conditions, k)
			args = append(args, v)
		}
		stmt += " WHERE " + strings.Join(conditions, " AND ")
	}

	if len(b.orders) > 0 {
		stmt += " ORDER BY " + strings.Join(b.orders, ",")
	}

	return stmt, args
}

func (b *CassandraQueryBuilder[T]) First() (object T) {
	stmt, args := b.buildQuery()
	return CassandraQueryFinalizeFirst[T](b.session, stmt, args, &CassandraQueryLimit{Limit: 1})
}

func (b *CassandraQueryBuilder[T]) Fetch(queryOptions ...CassandraQueryOption) (objects []T, result CassandraQueryResult[T]) {
	stmt, args := b.buildQuery()
	qos := make([]CassandraQueryOption, 0)
	if b.limit > 0 {
		qos = append(qos, &CassandraQueryLimit{Limit: b.limit})
	}

	if b.nextId != "" {
		qos = append(qos, &CassandraQueryNext{Next: b.nextId})
	}

	qos = append(qos, queryOptions...)
	return CassandraQueryFinalize[T](b.session, stmt, args, qos...)
}

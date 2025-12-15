package models

import (
	"context"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/gocql/gocql"
	base62 "github.com/yetiz-org/goth-base62"
	"github.com/yetiz-org/goth-bytebuf"
	"github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-util/hex"
	"github.com/yetiz-org/goth-util/value"
)

var IDCodec = base62.FlipShiftEncoding

type Model interface {
	TableName() string
}

func NewValidationError(field string, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

type ValidationError struct {
	Field   string
	Message string
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s: %s", v.Field, v.Message)
}

type LenValidationError ValidationError

func (v *LenValidationError) Error() string {
	return v.Field
}

type Validatable interface {
	Validate() bool
}

type ModelSavePreHook interface {
	PreSave(ctx context.Context) error
}

type ModelSavePostHook interface {
	PostSave(ctx context.Context) error
}

type ModelDeletePreHook interface {
	PreDelete(ctx context.Context) error
}

type ModelDeletePostHook interface {
	PostDelete(ctx context.Context) error
}

type CassandraModelScan interface {
	Scan(iter *gocql.Iter) bool
}

type Scope Privileges

func (s Scope) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *Scope) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	var privileges []string
	if err := json.Unmarshal(data, &privileges); err != nil {
		return err
	}
	*s = Scope(privileges)
	return nil
}

func (s Scope) Validate(scope string) bool {
	for _, sc := range s {
		scopePrefix := strings.Split(sc, ":")[0]
		if scopePrefix != "" {
			if strings.HasPrefix(scope, scopePrefix) {
				return true
			}
		}
	}

	return false
}

func (s Scope) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	switch info.Type() {
	case gocql.TypeText, gocql.TypeVarchar, gocql.TypeAscii:
	default:
		return nil, gocql.ErrUnsupported
	}

	if s == nil {
		return nil, nil
	}

	return json.Marshal(s)
}

func (s Scope) UnmarshalCQL(info gocql.TypeInfo, body []byte) (err error) {
	switch info.Type() {
	case gocql.TypeText, gocql.TypeVarchar, gocql.TypeAscii:
	default:
		return gocql.ErrUnsupported
	}

	if len(body) == 0 {
		return nil
	}
	v := ""
	_ = json.Unmarshal(body, &v)
	if err := json.Unmarshal([]byte(v), s); err != nil {
		return err
	}

	return nil
}

func (s *Scope) UnmarshalJSON(body []byte) (err error) {
	if len(body) == 0 || string(body) == "\"\"" {
		return nil
	}

	v := string(body)
	if v[0] == '"' {
		if err := json.Unmarshal(body, &v); err != nil {
			return err
		}
	}

	var privileges []string
	if err := json.Unmarshal([]byte(v), &privileges); err != nil {
		return err
	}

	*s = privileges
	return nil
}

type CredentialId string

func (a CredentialId) AppId() string {
	s := string(a)
	if strings.HasPrefix(s, "ast-") {
		s = strings.TrimPrefix(s, "ast-")
	}
	decoded, _ := base64.RawURLEncoding.DecodeString(s)
	if len(decoded) <= 16 {
		return ""
	}

	return strings.ToUpper(hex.EncodeToString(decoded[:16]))
}

func (a CredentialId) Id() string {
	return string(a)
}

type CredentialType string
type Privileges []string

func (p Privileges) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return json.Marshal(p)
}

func (p *Privileges) Scan(value interface{}) error {
	if value == nil {
		*p = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}

	var privileges []string
	if err := json.Unmarshal(data, &privileges); err != nil {
		return err
	}
	*p = Privileges(privileges)
	return nil
}

func (p Privileges) Validate(privilege string) bool {
	for _, pri := range p {
		privilegePath := strings.Split(pri, ":")[0]
		if privilegePath != "" {
			if strings.HasPrefix(privilege, privilegePath) {
				return true
			}
		}
	}

	return false
}

func (p Privileges) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	switch info.Type() {
	case gocql.TypeText, gocql.TypeVarchar, gocql.TypeAscii:
	default:
		return nil, gocql.ErrUnsupported
	}

	if p == nil {
		return nil, nil
	}

	return json.Marshal(p)
}

func (p Privileges) UnmarshalCQL(info gocql.TypeInfo, body []byte) (err error) {
	switch info.Type() {
	case gocql.TypeText, gocql.TypeVarchar, gocql.TypeAscii:
	default:
		return gocql.ErrUnsupported
	}

	if len(body) == 0 {
		return nil
	}
	v := ""
	_ = json.Unmarshal(body, &v)
	if err := json.Unmarshal([]byte(v), p); err != nil {
		return err
	}

	return nil
}

func (p *Privileges) UnmarshalJSON(body []byte) (err error) {
	if len(body) == 0 || string(body) == "\"\"" {
		return nil
	}

	v := string(body)
	if v[0] == '"' {
		if err := json.Unmarshal(body, &v); err != nil {
			return err
		}
	}

	var privileges []string
	if err := json.Unmarshal([]byte(v), &privileges); err != nil {
		return err
	}

	*p = privileges
	return nil
}

type Metadata map[string]any

func (c Metadata) MarshalCQL(info gocql.TypeInfo) ([]byte, error) {
	switch info.Type() {
	case gocql.TypeText, gocql.TypeVarchar, gocql.TypeAscii:
	default:
		return nil, gocql.ErrUnsupported
	}

	if c == nil {
		return nil, nil
	}

	return json.Marshal(c)
}

func (c Metadata) UnmarshalCQL(info gocql.TypeInfo, body []byte) (err error) {
	switch info.Type() {
	case gocql.TypeText, gocql.TypeVarchar, gocql.TypeAscii:
	default:
		return gocql.ErrUnsupported
	}

	if len(body) == 0 {
		return nil
	}
	v := ""
	_ = json.Unmarshal(body, &v)
	if err := json.Unmarshal([]byte(v), c); err != nil {
		return err
	}

	return nil
}

func (c *Metadata) UnmarshalJSON(body []byte) (err error) {
	if len(body) == 0 || string(body) == "\"\"" {
		return nil
	}

	v := string(body)
	if v[0] == '"' {
		if err := json.Unmarshal(body, &v); err != nil {
			return err
		}
	}

	m := map[string]any{}
	if err := json.Unmarshal([]byte(v), &m); err != nil {
		return err
	}

	*c = m
	return nil
}

const QueryOptionFlag = "GOTH_QUERY_OPTION"

type QueryOption interface {
	Operate(query *gocql.Query)
}

func DefaultQueryOption(next string) []QueryOption {
	return []QueryOption{&QueryNext{Next: next}, &QueryLimit{Limit: 50}}
}

type QueryLimit struct {
	Limit int
}

func (q *QueryLimit) Operate(query *gocql.Query) {
	if q.Limit > 0 {
		if _, f := query.Context().Value(QueryOptionFlag).(map[string]any)["PageState"]; !f {
			query.PageState(nil)
		}

		query.PageSize(q.Limit)
	}
}

type QueryNext struct {
	Next string
}

func (q *QueryNext) Operate(query *gocql.Query) {
	if q.Next != "" {
		bs := buf.NewByteBuf(IDCodec.DecodeString(q.Next))
		l := bs.ReadInt32()
		query.Context().Value(QueryOptionFlag).(map[string]any)["PageState"] = true
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

func prepareQuery(query *gocql.Query, opts ...QueryOption) *gocql.Query {
	query = query.WithContext(context.WithValue(context.Background(), QueryOptionFlag, map[string]any{}))
	for _, opt := range opts {
		opt.Operate(query)
	}

	return query
}

func QueryFinalizeFirst[T Model](session *gocql.Session, stmt string, args []any, opts ...QueryOption) (object T) {
	query := prepareQuery(session.Query(stmt, args...), append(opts, &QueryLimit{Limit: 1})...)
	iter := query.Iter()
	defer CloseIter(iter)

	if instance, ok := scanNext[T](iter); ok {
		object = instance
	}

	return
}

func QueryFinalize[T Model](session *gocql.Session, stmt string, args []any, opts ...QueryOption) (objects []T, result QueryResult[T]) {
	query := prepareQuery(session.Query(stmt, args...), opts...)
	objects = make([]T, 0)
	iter := query.Iter()
	defer CloseIter(iter)

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

type QueryResult[T Model] struct {
	Count        int    `json:"count"`
	NextId       string `json:"next_id"`
	session      *gocql.Session
	queryOptions []QueryOption
	query        *gocql.Query
}

func (q *QueryResult[T]) Next(session ...*gocql.Session) (objects []T, result QueryResult[T]) {
	if q.query == nil || q.NextId == "" {
		return
	}

	newQueryOptions := make([]QueryOption, 0)
	for _, opt := range q.queryOptions {
		if _, ok := opt.(*QueryNext); ok {
			continue
		}

		newQueryOptions = append(newQueryOptions, opt)
	}

	sess := q.session
	if len(session) > 0 {
		sess = session[0]
	}

	newQueryOptions = append(newQueryOptions, &QueryNext{Next: q.NextId})
	return QueryFinalize[T](sess, q.query.Statement(), q.query.Values(), newQueryOptions...)
}

func CloseIter(iter *gocql.Iter) {
	if iter != nil {
		if err := iter.Close(); err != nil {
			kklogger.WarnJ("models:Helper.CloseIter#keyspaces!close_error", err.Error())
		}
	}
}

type QueryBuilder[T Model] struct {
	session      *gocql.Session
	fields       []string
	conditions   map[string]any
	orders       []string
	queryOptions []QueryOption
	limit        int
	nextId       string
}

func NewQueryBuilder[T Model](session *gocql.Session) *QueryBuilder[T] {
	return &QueryBuilder[T]{session: session, conditions: map[string]any{}}
}

func (b *QueryBuilder[T]) Fields(fields ...string) *QueryBuilder[T] {
	b.fields = fields
	return b
}

func (b *QueryBuilder[T]) Where(condition string, arg any) *QueryBuilder[T] {
	b.conditions[condition] = arg
	return b
}

func (b *QueryBuilder[T]) Order(order string) *QueryBuilder[T] {
	b.orders = append(b.orders, order)
	return b
}

func (b *QueryBuilder[T]) Limit(limit int) *QueryBuilder[T] {
	b.limit = limit
	return b
}

func (b *QueryBuilder[T]) Next(nextId string) *QueryBuilder[T] {
	b.nextId = nextId
	return b
}

func (b *QueryBuilder[T]) buildQuery() (stmt string, args []any) {
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

func (b *QueryBuilder[T]) First() (object T) {
	stmt, args := b.buildQuery()
	return QueryFinalizeFirst[T](b.session, stmt, args, &QueryLimit{Limit: 1})
}

func (b *QueryBuilder[T]) Fetch(queryOptions ...QueryOption) (objects []T, result QueryResult[T]) {
	stmt, args := b.buildQuery()
	qos := make([]QueryOption, 0)
	if b.limit > 0 {
		qos = append(qos, &QueryLimit{Limit: b.limit})
	}

	if b.nextId != "" {
		qos = append(qos, &QueryNext{Next: b.nextId})
	}

	qos = append(qos, queryOptions...)
	return QueryFinalize[T](b.session, stmt, args, qos...)
}

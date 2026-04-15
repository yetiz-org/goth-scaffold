package models

import (
	"context"
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gocql/gocql"
	base62 "github.com/yetiz-org/goth-base62"
	"github.com/yetiz-org/goth-util/hex"
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

// ValidateColumnLength validates string field lengths against the size constraint
// defined in gorm struct tags (e.g., `gorm:"size:255"`).
// Returns *LenValidationError if any field exceeds its declared size limit.
func ValidateColumnLength(model any) error {
	v := reflect.ValueOf(model)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}

		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		gormTag := field.Tag.Get("gorm")
		if gormTag == "" {
			continue
		}

		sizeStr := gormTagValue(gormTag, "size")
		if sizeStr == "" {
			continue
		}

		size, err := strconv.Atoi(sizeStr)
		if err != nil || size <= 0 {
			continue
		}

		var strValue string
		switch fieldValue.Kind() {
		case reflect.String:
			strValue = fieldValue.String()
		case reflect.Pointer:
			if fieldValue.IsNil() {
				continue
			}

			elem := fieldValue.Elem()
			if elem.Kind() == reflect.String {
				strValue = elem.String()
			} else {
				continue
			}

		default:
			continue
		}

		columnName := gormTagValue(gormTag, "column")
		if columnName == "" {
			columnName = field.Name
		}

		if utf8.RuneCountInString(strValue) > size {
			return &LenValidationError{
				Field:   columnName,
				Message: fmt.Sprintf("exceeds maximum length of %d", size),
			}
		}
	}

	return nil
}

func gormTagValue(tag string, key string) string {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		if after, ok := strings.CutPrefix(part, key+":"); ok {
			return after
		}
	}

	return ""
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

// TimeOfDay represents a time-only value stored as MySQL TIME.
type TimeOfDay struct {
	time.Time
}

// Scan implements sql.Scanner for MySQL TIME values.
func (t *TimeOfDay) Scan(value any) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case []byte:
		return t.parseString(string(v))
	case string:
		return t.parseString(v)
	default:
		return fmt.Errorf("invalid time value: %T", value)
	}
}

func (t *TimeOfDay) parseString(value string) error {
	if value == "" {
		t.Time = time.Time{}
		return nil
	}

	parsed, err := time.Parse("15:04:05", value)
	if err != nil {
		return err
	}

	t.Time = parsed
	return nil
}

// Value implements driver.Valuer for MySQL TIME values.
func (t TimeOfDay) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return "00:00:00", nil
	}

	return t.Time.Format("15:04:05"), nil
}

// UnmarshalJSON implements json.Unmarshaler for MySQL TIME format "HH:MM:SS".
func (t *TimeOfDay) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "null" || s == "" {
		t.Time = time.Time{}
		return nil
	}

	return t.parseString(s)
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

func (s *Scope) UnmarshalCQL(info gocql.TypeInfo, body []byte) (err error) {
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
	if after, ok := strings.CutPrefix(s, "ast-"); ok {
		s = after
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

func (p *Privileges) UnmarshalCQL(info gocql.TypeInfo, body []byte) (err error) {
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

func (c Metadata) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}

	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return string(b), nil
}

func (c *Metadata) Scan(value any) error {
	if value == nil {
		*c = nil
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

	if len(data) == 0 {
		*c = nil
		return nil
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}

	*c = out
	return nil
}

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

func (c *Metadata) UnmarshalCQL(info gocql.TypeInfo, body []byte) (err error) {
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

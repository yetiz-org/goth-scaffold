package repositories

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gocql/gocql"
	"github.com/yetiz-org/goth-scaffold/app/connector/keyspaces"
	"github.com/yetiz-org/goth-scaffold/app/models"
)

type CassandraDefaultRepository[T models.Model] struct {
	session *gocql.Session
}

func (r *CassandraDefaultRepository[T]) try(query *gocql.Query) (result models.TryResult) {
	result.LastResult = map[string]any{}
	iter := query.Iter()
	iter.MapScan(result.LastResult)
	if result.LastResult != nil {
		if applied, f := result.LastResult["[applied]"]; f {
			result.LastApplied = applied.(bool)
		}
	}

	if !result.LastApplied {
		result.Error = models.ErrUniqueCreateNotApplied
		return
	}

	result.Error = iter.Close()
	return
}

func (r *CassandraDefaultRepository[T]) TableName() string {
	model := *new(T)
	return model.TableName()
}

func (r *CassandraDefaultRepository[T]) Session() *gocql.Session {
	return r.session
}

func (r *CassandraDefaultRepository[T]) QueryBuilder() *models.QueryBuilder[T] {
	return models.NewQueryBuilder[T](r.Session())
}

func (r *CassandraDefaultRepository[T]) SaveQuery(entity T) (stmt string, args []any) {
	stmt = "INSERT INTO " + entity.TableName() + "("
	refV := reflect.ValueOf(entity)
	if refV.Kind() == reflect.Ptr {
		refV = refV.Elem()
	}

	refVFCount := refV.NumField()
	for i, lastArgsLength := 0, 0; i < refVFCount; i++ {
		refVF := refV.Field(i)
		if !refVF.CanSet() {
			continue
		}

		cqlFieldName := refV.Type().Field(i).Tag.Get("cql")
		if cqlFieldName == "" {
			continue
		}

		if refVF.CanInt() {
			iv := refVF.Int()
			if cqlFieldName == "updated_at" && iv == 0 {
				args = append(args, time.Now().Unix())
			} else {
				args = append(args, iv)
			}
		} else if refVF.CanFloat() {
			args = append(args, refVF.Float())
		} else if refVF.Kind() == reflect.Bool {
			args = append(args, refVF.Bool())
		} else if refVF.Kind() == reflect.String {
			args = append(args, refVF.String())
		} else {
			args = append(args, refVF.Interface())
		}

		if cal := len(args); cal > lastArgsLength {
			if lastArgsLength == 0 {
				stmt += cqlFieldName
			} else {
				stmt = strings.Join([]string{stmt, cqlFieldName}, ",")
			}

			lastArgsLength = cal
		}
	}

	stmt += ") VALUES ("
	for i := range args {
		if i == 0 {
			stmt += "?"
		} else {
			stmt = strings.Join([]string{stmt, "?"}, ",")
		}
	}

	stmt += ")"
	if ttl, ok := any(entity).(models.CassandraModelTTL); ok {
		stmt = fmt.Sprintf("%s USING TTL %d", stmt, ttl.EntityTTL())
		if timestamp, ok := any(entity).(models.CassandraModelTimestamp); ok && timestamp.EntityTimestamp() > 0 {
			stmt = fmt.Sprintf("%s AND TIMESTAMP = %d", stmt, timestamp.EntityTimestamp())
		}
	} else {
		if timestamp, ok := any(entity).(models.CassandraModelTimestamp); ok && timestamp.EntityTimestamp() > 0 {
			stmt = fmt.Sprintf("%s USING TIMESTAMP = %d", stmt, timestamp.EntityTimestamp())
		}
	}

	return
}

func (r *CassandraDefaultRepository[T]) DeleteQuery(entity T) (stmt string, args []any) {
	modelMetadata, found := keyspaces.Writer().ColumnsMetadata()[entity.TableName()]
	if !found {
		return
	}

	stmt = "DELETE FROM " + entity.TableName() + " WHERE "
	refV := reflect.ValueOf(entity)
	if refV.Kind() == reflect.Ptr {
		refV = refV.Elem()
	}

	refVFCount := refV.NumField()
	for i, lastArgsLength := 0, 0; i < refVFCount; i++ {
		refVF := refV.Field(i)
		if !refVF.CanSet() {
			continue
		}

		cqlFieldName := refV.Type().Field(i).Tag.Get("cql")
		if cqlFieldName == "" {
			continue
		}

		if modelMetadata.Columns[cqlFieldName].Kind != "partition_key" && modelMetadata.Columns[cqlFieldName].Kind != "clustering" {
			continue
		}

		if refVF.CanInt() {
			args = append(args, refVF.Int())
		} else if refVF.CanFloat() {
			args = append(args, refVF.Float())
		} else if refVF.Kind() == reflect.Bool {
			args = append(args, refVF.Bool())
		} else if refVF.Kind() == reflect.String {
			args = append(args, refVF.String())
		} else {
			args = append(args, refVF.Interface())
		}

		if cal := len(args); cal > lastArgsLength {
			if lastArgsLength == 0 {
				stmt += fmt.Sprintf(" %s=?", cqlFieldName)
			} else {
				stmt = strings.Join([]string{stmt, fmt.Sprintf(" %s=?", cqlFieldName)}, " AND")
			}

			lastArgsLength = cal
		}
	}

	return
}

func (r *CassandraDefaultRepository[T]) UniqueCreate(entity T) error {
	if m, ok := any(entity).(models.ModelSavePreHook); ok {
		if err := m.PreSave(context.Background()); err != nil {
			return err
		}
	}

	stmt, args := r.SaveQuery(entity)
	if strings.Contains(strings.ToUpper(stmt), "USING TTL") {
		return errors.New("cassandra TTL is not supported when IF NOT EXISTS")
	}

	stmt = strings.TrimSuffix(strings.TrimSuffix(stmt, " "), ";")
	err := r.try(r.session.Query(fmt.Sprintf("%s IF NOT EXISTS;", stmt), args...)).Error
	if err == nil {
		if m, ok := any(entity).(models.ModelSavePostHook); ok {
			if err := m.PostSave(context.Background()); err != nil {
				return err
			}
		}
	}

	return err
}

func (r *CassandraDefaultRepository[T]) Save(entity T) error {
	if m, ok := any(entity).(models.ModelSavePreHook); ok {
		if err := m.PreSave(context.Background()); err != nil {
			return err
		}
	}

	stmt, args := r.SaveQuery(entity)
	err := r.session.Query(stmt, args...).Exec()
	if err == nil {
		if m, ok := any(entity).(models.ModelSavePostHook); ok {
			if err := m.PostSave(context.Background()); err != nil {
				return err
			}
		}
	}

	return err
}

func (r *CassandraDefaultRepository[T]) Delete(entity T) error {
	if m, ok := any(entity).(models.ModelDeletePreHook); ok {
		if err := m.PreDelete(context.Background()); err != nil {
			return err
		}
	}

	stmt, args := r.DeleteQuery(entity)
	if stmt == "" {
		return models.ErrModelMetadataNotFound
	}

	err := r.session.Query(stmt, args...).Exec()
	if err == nil {
		if m, ok := any(entity).(models.ModelDeletePostHook); ok {
			if err := m.PostDelete(context.Background()); err != nil {
				return err
			}
		}
	}

	return err
}

func NewCassandraDefaultRepository[T models.Model](session *gocql.Session) *CassandraDefaultRepository[T] {
	return &CassandraDefaultRepository[T]{
		session: session,
	}
}

package gosql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/frozenpine/pool"
	"github.com/valyala/bytebufferpool"
	"golang.org/x/exp/slog"
)

type TableDefine[T any] struct {
	schemaName string
	tableName  string
	columns    []*ColumnDefine
	dataType   reflect.Type
	pool       *pool.StructPool[T]
}

func (tbl *TableDefine[T]) filtedColumns(columns ...string) []*ColumnDefine {
	if len(columns) == 0 {
		return tbl.columns
	}

	var defines []*ColumnDefine

	for _, name := range columns {
		for _, col := range tbl.columns {
			if col.columnName == name {
				defines = append(defines, col)
				break
			}
		}
	}

	return defines
}

func (tbl *TableDefine[T]) setColumnValuesFn(columns ...string) func(interface{}) []interface{} {
	defines := tbl.filtedColumns(columns...)

	return func(v interface{}) []interface{} {
		if v == nil {
			return nil
		}

		basePtr := reflect.Indirect(reflect.ValueOf(v)).Addr().Pointer()
		colPtr := make([]interface{}, len(defines))

		for idx, col := range defines {
			colPtr[idx] = reflect.NewAt(
				col.fieldType, unsafe.Pointer(basePtr+col.fieldOffset),
			).Interface()
		}

		return colPtr
	}
}

func (tbl *TableDefine[T]) getColumnValuesFn(columns ...string) func(interface{}) []interface{} {
	defines := tbl.filtedColumns(columns...)

	return func(v interface{}) []interface{} {
		if v == nil {
			return nil
		}

		basePtr := reflect.Indirect(reflect.ValueOf(v)).Addr().Pointer()
		values := make([]interface{}, len(defines))

		for idx, col := range defines {
			values[idx] = reflect.Indirect(reflect.NewAt(
				col.fieldType, unsafe.Pointer(basePtr+col.fieldOffset),
			)).Interface()
		}

		return values
	}
}

func (tbl *TableDefine[T]) GetQueryTemplate(filter []*FilterDefine, columns ...string) string {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	defines := tbl.filtedColumns(columns...)
	colList := make([]string, len(defines))

	for idx, col := range defines {
		colList[idx] = col.columnName
	}

	buff.WriteString("SELECT ")
	buff.WriteString(strings.Join(colList, ","))
	buff.WriteString(" FROM ")
	if tbl.schemaName != "" {
		buff.WriteString(tbl.schemaName)
		buff.WriteString(".")
	}
	buff.WriteString(tbl.tableName)
	if len(filter) > 0 {
		buff.WriteString(" WHERE ")

		for idx, f := range filter {
			if idx > 0 {
				buff.WriteString(" AND ")
			}
			buff.WriteString(f.columnName)
			buff.WriteString(" ")
			buff.WriteString(string(f.action))
			buff.WriteString(" $")
			buff.WriteString(strconv.Itoa(idx + 1))
		}
	}
	buff.WriteString(";")

	return buff.String()
}

func (tbl *TableDefine[T]) GetInsertTemplate(columns ...string) string {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	defines := tbl.filtedColumns(columns...)
	colList := make([]string, len(defines))
	argList := make([]string, len(defines))

	for idx, col := range defines {
		colList[idx] = col.columnName
		argList[idx] = "$" + strconv.Itoa(idx+1)
	}

	buff.WriteString("INSERT INTO ")
	if tbl.schemaName != "" {
		buff.WriteString(tbl.schemaName)
		buff.WriteString(".")
	}
	buff.WriteString(tbl.tableName)
	buff.WriteString("(")
	buff.WriteString(strings.Join(colList, ","))
	buff.WriteString(") VALUES (")
	buff.WriteString(strings.Join(argList, ","))
	buff.WriteString(");")

	return buff.String()
}

func (tbl *TableDefine[T]) GetDropTemplate() string {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	buff.WriteString("DROP TABLE IF EXISTS ")
	if tbl.schemaName != "" {
		buff.WriteString(tbl.schemaName)
		buff.WriteString(".")
	}
	buff.WriteString(tbl.tableName)
	buff.WriteString(";")

	return buff.String()
}

func (tbl *TableDefine[T]) GetCreateTemplate() string {
	buff := bytebufferpool.Get()
	defer bytebufferpool.Put(buff)

	buff.WriteString("CREATE TABLE IF NOT EXISTS ")
	if tbl.schemaName != "" {
		buff.WriteString(tbl.schemaName)
		buff.WriteString(".")
	}
	buff.WriteString(tbl.tableName)
	buff.WriteString(" (")
	for idx, col := range tbl.columns {
		if idx > 0 {
			buff.WriteString(",")
		}
		buff.WriteString(col.columnName + " " + col.columnType.Type())
	}
	buff.WriteString(");")

	switch dbMode {
	case Postgres:
	case Sqlite3:
	}

	return buff.String()
}

func (tbl *TableDefine[T]) CompileQuery(filters []*FilterDefine, columns ...string) func(context.Context) ([]interface{}, error) {
	sqlTpl := tbl.GetQueryTemplate(filters, columns...)

	filterValues := make([]interface{}, len(filters))
	for idx, f := range filters {
		filterValues[idx] = f.value
	}

	valueSetFn := tbl.setColumnValuesFn(columns...)
	db := dbPtr.Load()

	return func(ctx context.Context) ([]interface{}, error) {
		slog.Debug("executing sql", "sql", sqlTpl, "args", filterValues)
		r, e := db.QueryContext(ctx, sqlTpl, filterValues...)
		if e != nil {
			return nil, e
		}

		results := []interface{}{}

		for r.Next() {
			data := tbl.pool.GetEmptyData(true)

			if err := r.Scan(valueSetFn(data)...); err != nil {
				return nil, err
			}

			results = append(results, data)
		}

		return results, nil
	}
}

func (tbl *TableDefine[T]) CompileInsert(columns ...string) func(context.Context, *T) (sql.Result, error) {
	sqlTpl := tbl.GetInsertTemplate(columns...)
	valueFn := tbl.getColumnValuesFn(columns...)
	db := dbPtr.Load()

	return func(ctx context.Context, v *T) (sql.Result, error) {
		values := valueFn(v)
		slog.Debug("executing sql", "sql", sqlTpl, "args", values)
		return db.ExecContext(ctx, sqlTpl, values...)
	}
}

func (tbl *TableDefine[T]) CompileBatchInsert(columns ...string) func(context.Context, ...*T) ([]sql.Result, error) {
	sqlTpl := tbl.GetInsertTemplate(columns...)
	valueFn := tbl.getColumnValuesFn(columns...)
	db := dbPtr.Load()

	return func(ctx context.Context, rows ...*T) ([]sql.Result, error) {
		tx, err := db.Begin()
		if err != nil {
			return nil, err
		}

		stmt, err := tx.PrepareContext(ctx, sqlTpl)
		if err != nil {
			return nil, err
		}

		result := make([]sql.Result, len(rows))

		for idx, row := range rows {
			values := valueFn(row)
			slog.Debug("executing sql", "sql", sqlTpl, "args", values)
			r, e := stmt.Exec(values...)
			if e != nil {
				err := tx.Rollback()
				if err != nil {
					return nil, fmt.Errorf("%w & rollback failed due to: %w", e, err)
				} else {
					return nil, e
				}
			} else {
				result[idx] = r
			}
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		return result, nil
	}
}

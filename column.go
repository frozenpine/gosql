package gosql

import "reflect"

type ColumnType uint8

const (
	ColBlob ColumnType = iota
	ColString
	ColTinyUint
	ColUint
	ColBigUint
	ColTinyInt
	ColInt
	ColBigInt
	ColTimestamp
	ColDatetime
	ColDate
	ColSingle
	ColDouble
	ColDecimal
	ColBoolean
)

func (col ColumnType) sqliteType() string {
	switch col {
	case ColBlob:
		return "BLOB"
	case ColString:
		return "TEXT"
	case ColTinyUint, ColUint, ColBigUint, ColTinyInt, ColInt, ColBigInt:
		return "INTEGER"
	case ColTimestamp:
		return "INTEGER"
	case ColDate, ColDatetime:
		return "TEXT"
	case ColSingle, ColDouble:
		return "REAL"
	case ColDecimal:
		return "NUMERIC"
	case ColBoolean:
		return "INTEGER"
	default:
		return "TEXT"
	}
}

func (col ColumnType) Type() string {
	switch dbMode {
	case Sqlite3:
		return col.sqliteType()
	case Postgres:
		return "VARCHAR"
	default:
		return "VARCHAR"
	}
}

type ColumnDefine struct {
	fieldOffset uintptr
	fieldType   reflect.Type
	columnName  string
	columnType  ColumnType
	isPrimary   bool
	isIndex     bool
	isUnique    bool
}

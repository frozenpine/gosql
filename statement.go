package gosql

type statementType uint8

//go:generate stringer -type statementType -linecomment
const (
	DDL statementType = iota // DDL
	DQL                      // DQL
	DML                      // DML
	DCL                      // DCL
	TCL                      // TCL
)

type StatementDefine struct {
	columns []Column
	filters []Filter
}

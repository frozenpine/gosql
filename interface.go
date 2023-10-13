package gosql

import "fmt"

type Result interface {
	IsSuccess() bool
	GetError() error
	GetAffectRows() int
}

type Filter interface {
	And(Filter) Filter
	Or(Filter) Filter
	Equal() Filter
	In() Filter
	Not() Filter

	fmt.Stringer
}

type Statement interface {
	Select(...string) Statement
	Insert(...string) Statement
	Update(...string) Statement
	FilterBy(...Filter) Statement
	Join(Statement, ...string) Statement

	GetType() statementType
	Execute(...any) Result
	Commit() Result

	fmt.Stringer
}

type Column interface {
	GetName() string
	GetType() string
}

type Table interface {
	Statement

	GetName() string
	GetColumns() []Column

	Create() Statement
	Drop() Statement
	WithTransaction() Statement
}

type View interface {
	Table
}

package gosql

import (
	"database/sql"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

type mode string

const (
	Postgres mode = "pgx"
	Sqlite3  mode = "sqlite3"
)

var (
	dbPtr  atomic.Pointer[sql.DB]
	dbMode mode

	modeList  = []mode{Postgres, Sqlite3}
	driverMap = stringMap{
		"postgres": "pgx",
		"pg":       "pgx",
		"sqlite":   "sqlite3",
	}

	connPattern = regexp.MustCompile(
		"(?P<proto>(?:" + strings.Join(
			append(
				*(*[]string)(unsafe.Pointer(&modeList)),
				driverMap.Keys()...,
			), "|") + "))://(?P<value>.+)",
	)
	protoIdx = connPattern.SubexpIndex("proto")
	valueIdx = connPattern.SubexpIndex("value")

	tableCache sync.Map
)

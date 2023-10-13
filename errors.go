package gosql

import "errors"

var (
	ErrInvalidDBConn = errors.New("invalid db conn string")
	ErrDBInit        = errors.New("database not initialized")
	ErrOpenDB        = errors.New("open database failed")
	ErrBackupDB      = errors.New("backup database failed")
	ErrDBMode        = errors.New("invalid db mode")
)

package sqlite3

import (
	"context"
	"database/sql"

	"github.com/frozenpine/gosql"
	"github.com/mattn/go-sqlite3"
)

func Sqlite3Backup(conn *sql.DB, dump string) error {
	if dump == "" {
		return nil
	}

	destDB, err := sql.Open("sqlite3", dump)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	destConn, err := destDB.Conn(ctx)
	if err != nil {
		return err
	}

	srcConn, err := conn.Conn(ctx)
	if err != nil {
		return err
	}

	return destConn.Raw(func(dest any) error {
		return srcConn.Raw(func(source any) error {
			dst, _ := dest.(*sqlite3.SQLiteConn)

			src, ok := source.(*sqlite3.SQLiteConn)
			if !ok {
				return gosql.ErrDBMode
			}

			backup, err := dst.Backup("main", src, "main")
			if err != nil {
				return err
			}

			backup.Step(-1)
			return backup.Finish()
		})
	})
}

package util

import (
	"context"
	"errors"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

// ConnectMySQL set the default sql db for transaction
func ConnectMySQL(dsn string) (err error) {
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
	}
	if dsn == "" {
		return errors.New("please set dsn by env MYSQL_DSN or manually setting")
	}

	if db, err = sqlx.Open("mysql", dsn); err != nil {
		return err
	}

	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

// Tx run functions with transaction
func Tx(ctx context.Context, fns ...func(tx *sqlx.Tx) (err error)) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	for _, fn := range fns {
		if err := fn(tx); err != nil {
			tx.Rollback()
			return err
		}

		if err := ctx.Err(); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

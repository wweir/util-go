package mysql

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/jmoiron/sqlx"
)

// MySQL is a wrap of sqlx DB
type MySQL struct {
	*sqlx.DB
}

var defaultDB *MySQL

// DB return the raw DB of sqlx
func DB() *sqlx.DB {
	return defaultDB.DB
}

// InitMySQL set the default sql db for transaction
func InitMySQL(dsn, dbName string) (err error) {
	defaultDB, err = ConnectMySQL(dsn, dbName)
	return
}

// ConnectMySQL connect to a MySQL database
func ConnectMySQL(dsn, dbName string) (*MySQL, error) {
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
	}
	if dsn == "" {
		return nil, errors.New("please set dsn by env MYSQL_DSN or manually setting")
	}

	if strings.IndexByte(dsn, '/') == -1 {
		dsn += "/" + dbName
	}
	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	conf.ParseTime = true
	if dbName != "" {
		conf.DBName = dbName
	}

	db, err := sqlx.Open("mysql", conf.FormatDSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+conf.DBName)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &MySQL{DB: db}, nil
}

// Tx run functions in default DB with transaction
func Tx(ctx context.Context, fns ...func(tx *sqlx.Tx) (err error)) error {
	return defaultDB.Tx(ctx, fns...)
}

// Tx run functions with transaction
func (db *MySQL) Tx(ctx context.Context, fns ...func(tx *sqlx.Tx) (err error)) error {
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

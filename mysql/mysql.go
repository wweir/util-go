package mysql

import (
	"context"
	"errors"
	"os"
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
func InitMySQL(dsn string) (err error) {
	defaultDB, err = ConnectMySQL(dsn)
	return
}

// ConnectMySQL connect to a MySQL database
func ConnectMySQL(dsn string) (*MySQL, error) {
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
	}
	if dsn == "" {
		return nil, errors.New("please set dsn by env MYSQL_DSN or manually setting")
	}

	db, err := sqlx.Open("mysql", dsn)
	if err == nil {
		err = db.Ping()
	}
	if err != nil {
		conf, e := mysql.ParseDSN(dsn)
		if e != nil {
			return nil, e
		} else if conf.DBName == "" {
			return nil, err
		}

		var dbName string
		var params map[string]string
		dbName, conf.DBName = conf.DBName, ""
		params, conf.Params = conf.Params, map[string]string{}
		if db, err = sqlx.Open("mysql", conf.FormatDSN()); err != nil {
			return nil, err
		}

		// create database if not open fail with the database: dbName
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = db.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+dbName)
		if err != nil {
			db.Close()
			return nil, err
		}
		db.Close()

		conf.DBName = dbName
		conf.Params = params
		if db, err = sqlx.Open("mysql", conf.FormatDSN()); err != nil {
			return nil, err
		}
		if err = db.Ping(); err != nil {
			return nil, err
		}
	}

	db.SetMaxOpenConns(256)
	db.SetConnMaxLifetime(20 * time.Second)
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

		select {
		case <-ctx.Done():
			if err := ctx.Err(); err != nil {
				tx.Rollback()
				return err
			}
		default:
		}
	}

	return tx.Commit()
}

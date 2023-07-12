package sql

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Interface interface {
	// BeginTransaction returns a Transaction for manual Commit and Rollback
	// For auto Commit and Rollback use InTransaction
	BeginTransaction() (Transaction, error)
	GetQuerier() Querier
	// InTransaction perform all the operations given in TransactionFunc and do a Commit if nil is returned.
	// Otherwise it will do a auto Rollback
	// Injected Context inside the TransactionFunc has the original Transaction object which can be accessible via
	// FromContext function
	InTransaction(context.Context, TransactionFunc) error
}

type TransactionFunc func(context.Context) error

type Querier interface {
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Preparex(query string) (*sqlx.Stmt, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type Transaction interface {
	Querier
	Commit() error
	Rollback() error
}

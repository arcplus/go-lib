package mysql

import (
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/arcplus/go-lib/errs"
)

var (
	In                = sqlx.In
	Get               = sqlx.Get
	GetContext        = sqlx.GetContext
	Select            = sqlx.Select
	SelectContext     = sqlx.SelectContext
	Named             = sqlx.Named
	NamedExec         = sqlx.NamedExec
	NamedExecContext  = sqlx.NamedExecContext
	NamedQuery        = sqlx.NamedQuery
	NamedQueryContext = sqlx.NamedQueryContext
)

// MySQLErr try conver mysql err to *mysql.MySQLError
func MySQLErr(err error) *mysql.MySQLError {
	if err == nil {
		return nil
	}
	if e, ok := err.(*mysql.MySQLError); ok {
		return e
	}
	return nil
}

// IsNoRowsErr
func IsNoRowsErr(err error) bool {
	return err == sql.ErrNoRows
}

const (
	ER_DUP_ENTRY = 1062
)

// IsDupErr check if mysql error is ER_DUP_ENTRY
// https://github.com/VividCortex/mysqlerr
func IsDupErr(err error) bool {
	e := MySQLErr(err)
	return e != nil && e.Number == ER_DUP_ENTRY
}

// IsChanged checks if result.RowsAffected is 0
func IsChanged(result sql.Result, err error) error {
	if err != nil {
		return err
	}

	aff, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if aff == 0 {
		return errs.New(errs.CodeNotFound, "RowsAffected is 0")
	}

	return nil
}

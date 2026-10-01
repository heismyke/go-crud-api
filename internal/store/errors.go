package store

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Sentinel errors. Stores return these (possibly wrapped with %w) so handlers
// can pick an HTTP status with errors.Is without knowing anything about SQL.
var (
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
	ErrNoCopies        = errors.New("no copies available")
	ErrLoanLimit       = errors.New("member has too many active loans")
	ErrAlreadyReturned = errors.New("loan already returned")
	ErrNotImplemented  = errors.New("not implemented")
)

// Postgres reports constraint failures with SQLSTATE codes.
// Full list: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgForeignKeyViolation = "23503"
	pgUniqueViolation     = "23505"
	pgCheckViolation      = "23514"
)

func pgErrCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

func isForeignKeyViolation(err error) bool { return pgErrCode(err) == pgForeignKeyViolation }
func isUniqueViolation(err error) bool     { return pgErrCode(err) == pgUniqueViolation }
func isCheckViolation(err error) bool      { return pgErrCode(err) == pgCheckViolation }

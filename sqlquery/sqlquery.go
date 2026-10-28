// Package sqlquery runs sqlmsg queries on a database/sql database, and runs
// functions inside transactions. It's tested with Postgres, through pgx's
// database/sql driver (github.com/jackc/pgx/v5/stdlib), which the
// composition root imports and opens. Other databases aren't supported yet.
//
// With pgx, each Postgres type arrives as one sqlmsg.Kind:
//
//   - smallint, integer, bigint: Int
//   - real, double precision: Float
//   - boolean: Bool
//   - timestamp, timestamptz, date: Time
//   - bytea, json, jsonb: Bytes
//   - text, varchar, numeric, uuid, interval, time, arrays and the rest:
//     Text, in Postgres's text format. Parse numeric with a decimal library,
//     so no precision is lost.
package sqlquery

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ianflanagan1/go-ports/sqlmsg"
)

// ErrUnsupportedType is returned for a column whose driver type has no
// sqlmsg.Kind.
var ErrUnsupportedType = errors.New("unsupported column type")

// DB runs queries on a database.
type DB struct {
	db *sql.DB
}

// New returns a DB that runs queries on db.
func New(db *sql.DB) DB {
	return DB{db: db}
}

// Query runs q and returns its rows.
func (d DB) Query(ctx context.Context, q sqlmsg.Query) (sqlmsg.Rows, error) {
	return query(ctx, d.db, q)
}

// Exec runs q, which returns no rows.
func (d DB) Exec(ctx context.Context, q sqlmsg.Query) (sqlmsg.Result, error) {
	return exec(ctx, d.db, q)
}

// InTx runs fn inside a transaction. It commits when fn returns nil, and
// rolls back when fn returns an error or panics.
func (d DB) InTx(ctx context.Context, fn func(sqlmsg.Tx) error) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning a transaction: %w", err)
	}
	done := false
	defer func() {
		if !done {
			_ = tx.Rollback() // fn panicked; the panic carries on.
		}
	}()
	err = fn(txQueries{ctx: ctx, tx: tx})
	done = true
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(err, fmt.Errorf("rolling back: %w", rollbackErr))
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing: %w", err)
	}
	return nil
}

// txQueries runs queries inside a transaction. It keeps the transaction's
// context, which database/sql's Tx does too, so sqlmsg.Tx needs none.
type txQueries struct {
	ctx context.Context
	tx  *sql.Tx
}

func (t txQueries) Query(q sqlmsg.Query) (sqlmsg.Rows, error) { return query(t.ctx, t.tx, q) }

func (t txQueries) Exec(q sqlmsg.Query) (sqlmsg.Result, error) { return exec(t.ctx, t.tx, q) }

// runner is what *sql.DB and *sql.Tx have in common.
type runner interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func query(ctx context.Context, r runner, q sqlmsg.Query) (sqlmsg.Rows, error) {
	rows, err := r.QueryContext(ctx, q.SQL, args(q.Args)...)
	if err != nil {
		return sqlmsg.Rows{}, fmt.Errorf("querying: %w", err)
	}
	defer func() { _ = rows.Close() }() // rows.Err reports what matters.
	columns, err := rows.Columns()
	if err != nil {
		return sqlmsg.Rows{}, fmt.Errorf("reading the columns: %w", err)
	}
	result := sqlmsg.Rows{Columns: columns}
	for rows.Next() {
		values, err := scan(rows, len(columns))
		if err != nil {
			return sqlmsg.Rows{}, err
		}
		result.Values = append(result.Values, values)
	}
	if err := rows.Err(); err != nil {
		return sqlmsg.Rows{}, fmt.Errorf("reading the rows: %w", err)
	}
	return result, nil
}

func exec(ctx context.Context, r runner, q sqlmsg.Query) (sqlmsg.Result, error) {
	result, err := r.ExecContext(ctx, q.SQL, args(q.Args)...)
	if err != nil {
		return sqlmsg.Result{}, fmt.Errorf("executing: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return sqlmsg.Result{}, fmt.Errorf("counting the rows affected: %w", err)
	}
	return sqlmsg.Result{RowsAffected: affected}, nil
}

func scan(rows *sql.Rows, n int) ([]sqlmsg.Value, error) {
	raw := make([]any, n)
	targets := make([]any, n)
	for i := range raw {
		targets[i] = &raw[i]
	}
	if err := rows.Scan(targets...); err != nil {
		return nil, fmt.Errorf("reading a row: %w", err)
	}
	values := make([]sqlmsg.Value, n)
	for i, v := range raw {
		value, err := toValue(v)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// toValue converts a value as the driver returns it.
func toValue(v any) (sqlmsg.Value, error) {
	switch v := v.(type) {
	case nil:
		return sqlmsg.NullValue(), nil
	case int64:
		return sqlmsg.IntValue(v), nil
	case float64:
		return sqlmsg.FloatValue(v), nil
	case string:
		return sqlmsg.TextValue(v), nil
	case []byte:
		return sqlmsg.BytesValue(v), nil
	case bool:
		return sqlmsg.BoolValue(v), nil
	case time.Time:
		return sqlmsg.TimeValue(v), nil
	}
	return sqlmsg.Value{}, fmt.Errorf("%w: %T", ErrUnsupportedType, v)
}

func args(values []sqlmsg.Value) []any {
	converted := make([]any, len(values))
	for i, v := range values {
		converted[i] = argument(v)
	}
	return converted
}

func argument(v sqlmsg.Value) any {
	switch v.Kind {
	case sqlmsg.Null:
		return nil
	case sqlmsg.Int:
		return v.Int
	case sqlmsg.Float:
		return v.Float
	case sqlmsg.Text:
		return v.Text
	case sqlmsg.Bytes:
		return v.Bytes
	case sqlmsg.Bool:
		return v.Bool
	case sqlmsg.Time:
		return v.Time
	}
	return nil
}

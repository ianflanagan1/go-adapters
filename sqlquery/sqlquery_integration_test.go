//go:build integration

package sqlquery_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/ianflanagan1/go-adapters/sqlquery"
	"github.com/ianflanagan1/go-ports/sqlmsg"
)

// open connects to the Postgres at GO_ADAPTERS_POSTGRES_URL, such as one
// started with:
//
//	docker run --rm -d -e POSTGRES_PASSWORD=test -p 55432:5432 postgres:17
//	export GO_ADAPTERS_POSTGRES_URL=postgres://postgres:test@localhost:55432/postgres
func open(t *testing.T) sqlquery.DB {
	t.Helper()
	url := os.Getenv("GO_ADAPTERS_POSTGRES_URL")
	if url == "" {
		t.Skip("GO_ADAPTERS_POSTGRES_URL isn't set")
	}
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	d := sqlquery.New(db)
	for _, stmt := range []string{
		"DROP TABLE IF EXISTS sqlquery_test",
		"CREATE TABLE sqlquery_test (id bigint PRIMARY KEY, name text, score double precision, data bytea, ok boolean, at timestamptz, amount numeric)",
	} {
		if _, err := d.Exec(context.Background(), sqlmsg.Query{SQL: stmt}); err != nil {
			t.Fatal(err)
		}
	}
	return d
}

func TestEachKindRoundTrips(t *testing.T) {
	d := open(t)
	ctx := context.Background()
	at := time.Date(2026, time.October, 28, 9, 30, 0, 0, time.UTC)
	insert := sqlmsg.Query{
		SQL: "INSERT INTO sqlquery_test VALUES ($1, $2, $3, $4, $5, $6, $7)",
		Args: []sqlmsg.Value{
			sqlmsg.IntValue(1), sqlmsg.TextValue("a"), sqlmsg.FloatValue(1.5),
			sqlmsg.BytesValue([]byte{1, 2}), sqlmsg.BoolValue(true), sqlmsg.TimeValue(at), sqlmsg.TextValue("12.345"),
		},
	}
	if result, err := d.Exec(ctx, insert); err != nil || result.RowsAffected != 1 {
		t.Fatalf("got %+v, %v", result, err)
	}
	rows, err := d.Query(ctx, sqlmsg.Query{SQL: "SELECT id, name, score, data, ok, at, amount, NULL::text AS missing FROM sqlquery_test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows.Values) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows.Values))
	}
	row := rows.Values[0]
	if row[0].Int != 1 || row[1].Text != "a" || row[2].Float != 1.5 || string(row[3].Bytes) != "\x01\x02" ||
		!row[4].Bool || !row[5].Time.Equal(at) || row[6].Text != "12.345" || row[7].Kind != sqlmsg.Null {
		t.Fatalf("got %+v", row)
	}
	if i, ok := rows.Column("amount"); !ok || i != 6 {
		t.Fatalf("got column %d, %v", i, ok)
	}
}

func TestATransactionCommitsWhenItsFunctionSucceeds(t *testing.T) {
	d := open(t)
	ctx := context.Background()
	err := d.InTx(ctx, func(tx sqlmsg.Tx) error {
		_, err := tx.Exec(sqlmsg.Query{SQL: "INSERT INTO sqlquery_test (id) VALUES (1)"})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	assertRows(t, d, 1)
}

func TestATransactionRollsBackWhenItsFunctionFails(t *testing.T) {
	d := open(t)
	failure := errors.New("failed")
	err := d.InTx(context.Background(), func(tx sqlmsg.Tx) error {
		if _, err := tx.Exec(sqlmsg.Query{SQL: "INSERT INTO sqlquery_test (id) VALUES (1)"}); err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("got %v, want the function's error", err)
	}
	assertRows(t, d, 0)
}

func TestATransactionRollsBackWhenItsFunctionPanics(t *testing.T) {
	d := open(t)
	func() {
		defer func() { _ = recover() }()
		_ = d.InTx(context.Background(), func(tx sqlmsg.Tx) error {
			if _, err := tx.Exec(sqlmsg.Query{SQL: "INSERT INTO sqlquery_test (id) VALUES (1)"}); err != nil {
				return err
			}
			panic("boom")
		})
	}()
	assertRows(t, d, 0)
}

func TestAQueryInsideATransactionSeesItsWrites(t *testing.T) {
	d := open(t)
	err := d.InTx(context.Background(), func(tx sqlmsg.Tx) error {
		if _, err := tx.Exec(sqlmsg.Query{SQL: "INSERT INTO sqlquery_test (id) VALUES (1)"}); err != nil {
			return err
		}
		rows, err := tx.Query(sqlmsg.Query{SQL: "SELECT id FROM sqlquery_test"})
		if err == nil && len(rows.Values) != 1 {
			t.Errorf("got %d rows inside the transaction, want 1", len(rows.Values))
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOtherPostgresTypesArriveAsTextOrBytes(t *testing.T) {
	d := open(t)
	rows, err := d.Query(context.Background(), sqlmsg.Query{
		SQL: `SELECT ARRAY[1, 2], '00000000-0000-0000-0000-000000000001'::uuid, '1 hour'::interval, '{"a": 1}'::jsonb`,
	})
	if err != nil {
		t.Fatal(err)
	}
	row := rows.Values[0]
	if row[0].Text != "{1,2}" || row[1].Text != "00000000-0000-0000-0000-000000000001" ||
		row[2].Text != "01:00:00" || row[3].Kind != sqlmsg.Bytes || string(row[3].Bytes) != `{"a": 1}` {
		t.Fatalf("got %+v", row)
	}
}

func TestABadQueryIsAnError(t *testing.T) {
	d := open(t)
	if _, err := d.Query(context.Background(), sqlmsg.Query{SQL: "SELECT FROM nowhere"}); err == nil {
		t.Fatal("got no error")
	}
	if _, err := d.Exec(context.Background(), sqlmsg.Query{SQL: "DELETE FROM nowhere"}); err == nil {
		t.Fatal("got no error")
	}
}

func assertRows(t *testing.T, d sqlquery.DB, want int) {
	t.Helper()
	rows, err := d.Query(context.Background(), sqlmsg.Query{SQL: "SELECT id FROM sqlquery_test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows.Values) != want {
		t.Fatalf("got %d rows, want %d", len(rows.Values), want)
	}
}

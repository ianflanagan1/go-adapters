# go-adapters

Adapters shared by every project on [go-baseline](../go-baseline). Each package implements a port by doing I/O and makes no decisions, so it can be reviewed once and trusted everywhere. Import them only from a composition root (`cmd/`) or a project's own adapters; the shell declares the ports they satisfy.

| Package | Adapts | Satisfies a port like |
|---|---|---|
| `clock` | The system clock | `Now() time.Time` |
| `env` | The process environment | `Lookup(key string) (string, bool)` |
| `random` | `crypto/rand` (`Secure`) | `io.Reader` |
| `files` | A directory, through `os.Root`, so paths can't escape it | `ReadFile(name string) ([]byte, error)`, `WriteFile(name string, content []byte) error` |
| `httpsend` | HTTP, through `net/http` | `Send(ctx context.Context, req httpmsg.Request) (httpmsg.Response, error)` |
| `sqlquery` | Postgres, through `database/sql` and pgx | `Query(ctx, sqlmsg.Query) (sqlmsg.Rows, error)`, `Exec(ctx, sqlmsg.Query) (sqlmsg.Result, error)`, `InTx(ctx, func(sqlmsg.Tx) error) error` |

The request, response, query and row types are in [go-ports](https://github.com/ianflanagan1/go-ports), a pure library, so a project's core can build and read them and its shell can declare ports in them.

`sqlquery` supports Postgres only; its package documentation lists how each Postgres type arrives. `InTx` runs a function inside a transaction: it commits when the function returns nil, and rolls back when it returns an error or panics.

## Testing

`devbox run -- task test` runs everything that needs nothing outside the process. `sqlquery`'s tests need Postgres:

```sh
docker run --rm -d -e POSTGRES_PASSWORD=test -p 55432:5432 postgres:17
GO_ADAPTERS_POSTGRES_URL=postgres://postgres:test@localhost:55432/postgres devbox run -- task integration
```

Add an adapter here when more than one project needs it. Keep it thin: when it needs a decision, the decision belongs in the caller's core.

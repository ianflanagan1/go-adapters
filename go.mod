module github.com/ianflanagan1/go-adapters

go 1.27

tool github.com/ianflanagan1/go-baseline/cmd/go-baseline

require (
	github.com/ianflanagan1/go-ports v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.11.0
)

require (
	github.com/ianflanagan1/go-baseline v0.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
)

// Until go-ports is published.
replace github.com/ianflanagan1/go-ports => ../go-ports

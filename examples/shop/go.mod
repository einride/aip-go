module github.com/ysomad/flop/examples/shop

go 1.25.0

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/jackc/pgx/v5 v5.10.0
	github.com/ysomad/flop v0.0.6
	github.com/ysomad/flop/sqflop v0.0.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.29.0 // indirect
)

replace github.com/ysomad/flop => ../..

replace github.com/ysomad/flop/sqflop => ../../sqflop

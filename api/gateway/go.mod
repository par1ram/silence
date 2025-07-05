module github.com/par1ram/silence/api/gateway

go 1.23.0

toolchain go1.23.2

require (
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/par1ram/silence/shared v0.0.0
	go.uber.org/zap v1.26.0
	golang.org/x/time v0.12.0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

replace github.com/par1ram/silence/shared => ../../shared

module github.com/par1ram/silence/api/auth

go 1.23.0

toolchain go1.23.2

require (
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/par1ram/silence/shared v0.0.0
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.39.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/par1ram/silence/shared => ../../shared

module github.com/par1ram/silence/api/gateway

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/par1ram/silence/shared v0.0.0
	go.uber.org/zap v1.26.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/par1ram/silence/shared => ../../shared

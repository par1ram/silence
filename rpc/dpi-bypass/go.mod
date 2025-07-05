module github.com/par1ram/silence/rpc/dpi-bypass

go 1.23

require (
	github.com/par1ram/silence/shared v0.0.0
	go.uber.org/zap v1.26.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/par1ram/silence/shared => ../../shared

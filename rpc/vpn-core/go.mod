module github.com/par1ram/silence/rpc/vpn-core

go 1.23.0

toolchain go1.23.2

require (
	github.com/par1ram/silence/shared v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.26.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20241231184526-a9ab2273dd10
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/josharian/native v1.1.0 // indirect
	github.com/mdlayher/genetlink v1.3.2 // indirect
	github.com/mdlayher/netlink v1.7.2 // indirect
	github.com/mdlayher/socket v0.5.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.zx2c4.com/wireguard v0.0.0-20231211153847-12269c276173 // indirect
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.39.0
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
)

replace github.com/par1ram/silence/shared => ../../shared

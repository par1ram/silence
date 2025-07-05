module github.com/par1ram/silence/rpc/notifications

go 1.23

require (
	google.golang.org/grpc v1.57.0
	google.golang.org/protobuf v1.31.0
	github.com/streadway/amqp v1.1.0
	github.com/par1ram/silence/shared v0.0.0
)

replace github.com/par1ram/silence/shared => ../../shared

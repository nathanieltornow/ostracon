module github.com/nathanieltornow/ostracon/shard

go 1.16

replace github.com/nathanieltornow/ostracon => ../

require (
	github.com/nathanieltornow/ostracon v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)

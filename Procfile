
seqshard1: go run cmd/startshard/startshard.go -isRoot -isSequencer -ipAddr localhost:4000
#recshard1: go run cmd/startshard/startshard.go -parentIpAddr localhost:4000 -ipAddr localhost:4001 -interval 100ms
#client1: go run cmd/client/client.go
#client2: go run cmd/client/client.go
#client3: go run cmd/client/client.go
#client4: go run cmd/client/client.go

recshard1: go run cmd/startrecordshard/startrecordshard.go -parentIpAddr localhost:4000 -ipAddr localhost:6000 -storagePath tmp/shard1
#recshard2: go run cmd/startrecordshard/startrecordshard.go -parentIpAddr localhost:4000 -ipAddr localhost:6001 -storagePath tmp/shard2

client1: go run cmd/client/client.go -parentIpAddr localhost:6000
client2: go run cmd/client/client.go -parentIpAddr localhost:6000

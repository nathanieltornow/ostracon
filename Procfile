
seqshard1: go run cmd/startshard/startshard.go -isRoot -isSequencer -ipAddr localhost:4000
recshard1: go run cmd/startshard/startshard.go -parentIpAddr localhost:4000 -ipAddr localhost:4001 -interval 10ms
client1: go run cmd/client/client.go
client2: go run cmd/client/client.go
#client3: go run cmd/client/client.go
#client4: go run cmd/client/client.go
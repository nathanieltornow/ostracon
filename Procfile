
seqshard1: go run cmd/startshard/startshard.go -isRoot -isSequencer -ipAddr localhost:4000
recshard1: go run cmd/startshard/startshard.go -parentIpAddr localhost:4000 -ipAddr localhost:4001
client: go run cmd/client/client.go
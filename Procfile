
seqshard1: go run cmd/shard/shard.go -isRoot -ipAddr localhost:4001 -color 0
seqshard2: go run cmd/shard/shard.go -ipAddr localhost:4000 -parentIpAddr localhost:4001 -interval 100us -color 1
seqshard3: go run cmd/shard/shard.go -ipAddr localhost:4002 -parentIpAddr localhost:4000 -interval 100us -color 2

recshard1: go run cmd/shard/shard.go -rec -parentIpAddr localhost:4000 -ipAddr localhost:6001 -storagePath tmp/shard2 -interval 100us
recshard2: go run cmd/shard/shard.go -rec -parentIpAddr localhost:4000 -ipAddr localhost:6002 -storagePath tmp/shard2 -interval 100us
recshard3: go run cmd/shard/shard.go -rec -parentIpAddr localhost:4002 -ipAddr localhost:6003 -storagePath tmp/shard2 -interval 100us
#recshard3: go run cmd/shard/shard.go -rec -parentIpAddr localhost:4000 -ipAddr localhost:6002 -storagePath tmp/shard2 -interval 100us

#client1: go run cmd/client/client.go -parentIpAddr localhost:6000
#client2: go run cmd/client/client.go -parentIpAddr localhost:6000
#client3: go run cmd/client/client.go -parentIpAddr localhost:6000
#client4: go run cmd/client/client.go -parentIpAddr localhost:6001
#client5: go run cmd/client/client.go -parentIpAddr localhost:6001
#client6: go run cmd/client/client.go -parentIpAddr localhost:6001
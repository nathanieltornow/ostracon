.PHONY: build
build:
	go build -o ostracon ostracon.go

start-cluster:
	go run cmd/start_cluster/start_cluster.go
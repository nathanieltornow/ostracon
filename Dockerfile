FROM golang:1.16

RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main cmd/shard/shard.go
ENTRYPOINT ["/app/main"]
EXPOSE 4000
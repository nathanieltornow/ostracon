#FROM golang:1.16
#
#RUN mkdir /app
#ADD shard /app
#WORKDIR /app/shard
#RUN go build -o main main.go
#
#CMD ["/app/shard/main"]
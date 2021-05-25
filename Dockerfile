FROM golang:1.16

ARG FLAGS

RUN mkdir /app
ADD . /app
WORKDIR /app
RUN go build -o main cmd/startshard/startshard.go

CMD /app/main $FLAGS
FROM golang:1.15.7-alpine3.13

RUN apk update -qq && apk add git && apk add --no-cache bash

WORKDIR /go/src/aapanavypar_service_updater

ADD . .

RUN go mod download

RUN go build -o main ./server/main.go

CMD ["./main"]

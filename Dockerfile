FROM golang:1.14.2-alpine3.11 AS Builder

RUN mkdir -p /go/fileServer
WORKDIR /go/fileServer

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build

#--------------------------------
FROM alpine:3.11.5

COPY --from=Builder /go/fileServer/fileServer /bin

ENTRYPOINT ["fileServer"]

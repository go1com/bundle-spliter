FROM golang:1.11

WORKDIR /go/src/github.com/go1com/bundle-spliter/
COPY    . /go/src/github.com/go1com/bundle-spliter/

RUN go get github.com/golang/dep/cmd/dep
RUN pwd; ${GOPATH}/bin/dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -o /app /go/src/github.com/go1com/bundle-spliter/cmd/main.go

FROM alpine:3.8
RUN apk add --no-cache ca-certificates
COPY --from=0 /app /app
ENTRYPOINT ["/app"]

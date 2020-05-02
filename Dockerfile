FROM golang:1.7.3 AS builder
WORKDIR /go/src/github.com/mettizik/gosonarci/
COPY src .
RUN GOOS=linux GOARCH=386 CGO_ENABLED=0 go build main.go

FROM alpine:3.9.6
WORKDIR /root/
COPY --from=builder /go/src/github.com/mettizik/gosonarci/main ./sqqg
ENTRYPOINT ["./sqqg"]

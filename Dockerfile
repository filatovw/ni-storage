FROM golang:1.12 as builder
COPY . $GOPATH/src/github.com/filatovw/ni-storage
WORKDIR $GOPATH/src/github.com/filatovw/ni-storage/
RUN go install -v ./...
RUN go build -o /usr/bin/api ./cmd/api


FROM debian:stretch
WORKDIR /usr/bin/
COPY --from=builder /usr/bin/api .
CMD ["/usr/bin/api", "-data-dir", "/var/lib/data"]
VOLUME ["/var/lib/data"]
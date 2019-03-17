FROM golang:1.12 as builder
COPY . /ni-storage
WORKDIR /ni-storage
RUN go mod download
RUN go build -o /usr/bin/api ./cmd/api

FROM debian:stretch
WORKDIR /usr/bin/
COPY --from=builder /usr/bin/api .
CMD ["/usr/bin/api", "-data-dir", "/var/lib/data"]
VOLUME ["/var/lib/data"]
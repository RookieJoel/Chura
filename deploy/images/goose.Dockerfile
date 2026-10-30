FROM golang:1.26-alpine AS builder
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.29.0

FROM alpine:3.22
COPY --from=builder /go/bin/goose /bin/goose
ENTRYPOINT ["/bin/goose"]

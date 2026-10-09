FROM golang:1.25-alpine AS builder
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.24.3

FROM alpine:3.20
COPY --from=builder /go/bin/goose /bin/goose
ENTRYPOINT ["/bin/goose"]
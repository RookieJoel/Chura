FROM golang:1.26-alpine AS builder
# Build goose inside a throwaway module so its dependencies can be bumped past
# goose's own pins (go install pkg@version always uses the upstream go.mod).
# golang.org/x/net >= v0.62.0 fixes CVE-2026-78661; goose v3.29.0 pins v0.61.0.
WORKDIR /build
RUN go mod init goose-build \
    && go get github.com/pressly/goose/v3/cmd/goose@v3.29.0 golang.org/x/net@v0.62.0 \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /go/bin/goose github.com/pressly/goose/v3/cmd/goose

FROM alpine:3.22
COPY --from=builder /go/bin/goose /bin/goose
ENTRYPOINT ["/bin/goose"]

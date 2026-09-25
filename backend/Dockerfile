FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/api ./cmd/api

FROM alpine:3.22

WORKDIR /app
COPY --from=builder /out/api ./api

EXPOSE 8080 9000
ENTRYPOINT ["/app/api"]

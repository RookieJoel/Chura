FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /out/project ./cmd/project && \
    go build -o /out/sprint ./cmd/sprint && \
    go build -o /out/analytics ./cmd/analytics && \
    go build -o /out/workitem ./cmd/workitem

FROM alpine:3.22

WORKDIR /app
COPY --from=builder /out/project ./project
COPY --from=builder /out/sprint ./sprint
COPY --from=builder /out/analytics ./analytics
COPY --from=builder /out/workitem ./workitem

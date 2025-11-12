# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-greeter .

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/mcp-greeter /app/mcp-greeter

ENTRYPOINT ["/app/mcp-greeter"]

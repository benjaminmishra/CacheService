FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o cache-service ./cmd/cache-service

FROM alpine:3.18
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/cache-service ./cache-service
EXPOSE 8080
ENV PORT=8080
CMD ["./cache-service"]

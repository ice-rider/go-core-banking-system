FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY server/go.mod server/go.sum* ./
RUN go mod download

COPY server/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api-gateway ./cmd/api-gateway

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/api-gateway /bin/api-gateway

EXPOSE 8080

ENTRYPOINT ["/bin/api-gateway"]

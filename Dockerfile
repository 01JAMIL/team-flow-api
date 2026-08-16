FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

RUN adduser -D -u 1000 appuser

WORKDIR /app
COPY --from=builder /app/bin/server .

USER appuser

EXPOSE 3700

CMD ["./server"]
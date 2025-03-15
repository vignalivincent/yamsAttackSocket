FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .
COPY static/ /app/static/

ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
# Stage 1: Build
FROM golang:latest AS builder 

WORKDIR /app

# Copy go mod and sum files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy entire source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-api ./cmd/api

# Stage 2: Runtime
FROM debian:bullseye-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Bangkok

# ✅ Copy built binary
COPY --from=builder /app/chat-api .

# ✅ Copy config folder (ตรง path จริง)
COPY --from=builder /app/pkg/configs ./pkg/configs

# ✅ Copy .env
COPY .env .env

RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["./chat-api"]

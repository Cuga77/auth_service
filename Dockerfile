FROM golang:alpine AS builder

WORKDIR /app
COPY . .

RUN apk add --no-cache git
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/auth-service .
COPY --from=builder /app/migrations ./migrations

RUN apk add --no-cache libc6-compat
RUN apk add --no-cache tzdata
ENV TZ=Europe/Moscow

EXPOSE 8081
CMD ["./auth-service"]
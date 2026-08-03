FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o enterprise-api .

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/enterprise-api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./enterprise-api"]

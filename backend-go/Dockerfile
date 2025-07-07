#Build stage
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o backend ./cmd/main.go

#Final stage
FROM golang:1.24

WORKDIR /app

COPY --from=builder /app/backend .

CMD ["./backend"]
# Build stage
FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./server

# Run stage
FROM scratch

WORKDIR /

COPY --from=builder /server /server

EXPOSE 8080

CMD ["./server"]
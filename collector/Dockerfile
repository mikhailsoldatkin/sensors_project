FROM golang:1.22 as builder

WORKDIR /app

RUN apt-get update

COPY go.mod go.sum ./

RUN go mod download

COPY main.go .
COPY proto/ /app/proto

RUN CGO_ENABLED=0 GOOS=linux go build -o collector .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/collector .

CMD ["./collector"]

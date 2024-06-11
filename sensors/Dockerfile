FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/main.go .
COPY pkg/api /app/pkg/api

RUN CGO_ENABLED=0 GOOS=linux go build -o sensors .

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/sensors .

CMD ["./sensors"]

FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /tinvest .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /tinvest .

RUN mkdir -p /app/data

CMD ["./tinvest"]

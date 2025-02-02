FROM golang:1.23.4-alpine3.20 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN apk add --no-cache git
RUN go mod download

COPY cmd/ ./cmd
COPY internal/ ./internal

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build -o /api-service cmd/app/main.go

FROM alpine:3.20

WORKDIR /app

COPY --from=build /api-service .

EXPOSE 8080

CMD ["./api-service"]

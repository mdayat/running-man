# syntax=docker/dockerfile:1
FROM alpine:3.21 AS base-alpine
WORKDIR /app

FROM golang:1.23.6-alpine3.21 AS base-go
WORKDIR /app

FROM base-go AS build
COPY go.mod go.sum ./
RUN go mod download
COPY configs configs
COPY internal internal
COPY repository repository
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o webhook

FROM base-alpine AS final
COPY --from=build /app/webhook .
COPY .env .
ENTRYPOINT ["/app/webhook"]
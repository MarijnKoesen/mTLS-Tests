## Basic build
FROM golang:1.22-bookworm AS base
WORKDIR /app
COPY ./go.mod ./
RUN go mod download -x
COPY ./ ./

## Build
FROM base AS build
ENV CGO_ENABLED=0
WORKDIR /app/server
RUN go build -o /server ./server.go

WORKDIR /app/client
RUN go build -o /client ./client.go

## Deploy
FROM alpine:latest as prod
WORKDIR /
COPY --from=build /server /server
COPY --from=build /client /client
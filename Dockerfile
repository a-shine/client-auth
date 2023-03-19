# syntax=docker/dockerfile:1

## Build
FROM golang:1.19 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /client-authentication

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /client-authentication /client-authentication

USER nonroot:nonroot

CMD ["/client-authentication"]
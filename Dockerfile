# syntax=docker/dockerfile:1

## Build
FROM golang:1.19 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /user-auth

## Deploy
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /user-auth /user-auth

USER nonroot:nonroot

CMD ["/user-auth"]
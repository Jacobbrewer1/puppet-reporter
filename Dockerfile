FROM golang:alpine as build
WORKDIR /build

COPY go.sum go.mod /build/

RUN go mod download
RUN go mod tidy

COPY . /build/
RUN go build -o application

FROM ubuntu:latest

COPY --from=build /build/application /usr/local/bin/application
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["application"]

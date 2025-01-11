FROM docker.io/golang:alpine as build

ARG APP_NAME=app
ENV APP_NAME=${APP_NAME}

WORKDIR /build

COPY . /build/

RUN go mod download
RUN go mod tidy

COPY . /build/
RUN go build -o application cmd/${APP_NAME}/main.go

FROM docker.io/ubuntu:latest

COPY --from=build /build/application /usr/local/bin/application
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["application"]

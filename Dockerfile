FROM docker.io/golang:alpine as build
WORKDIR /build

COPY . /build/

RUN go mod download
RUN go mod tidy

COPY . /build/
RUN cd ./cmd/ && go build -o application

FROM docker.io/ubuntu:latest

COPY --from=build /build/cmd/application /usr/local/bin/application
ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["application"]

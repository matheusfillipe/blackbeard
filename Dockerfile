FROM golang:1.18.5-buster
    MAINTAINER Matheus Fillipe <matheusfillipeag@gmail.com>

ARG version="v0.5.2"
RUN apt-get update && apt-get -y install curl libcurl4-openssl-dev wget pkg-config libc-dev gcc
ADD . /blb
WORKDIR /blb
RUN go build .
RUN mkdir -p libcurl-impersonate && \
    cd libcurl-impersonate && \
    wget "https://github.com/lwthiker/curl-impersonate/releases/download/${version}/libcurl-impersonate-${version}.x86_64-linux-gnu.tar.gz" -O libcurl-impersonate.tar.gz && \
    tar -xzf libcurl-impersonate.tar.gz
CMD LD_PRELOAD=libcurl-impersonate/libcurl-impersonate-chrome.so ./blackbeard -api

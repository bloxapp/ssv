#
# STEP 1: Prepare environment
#
FROM golang:1.19 AS preparer

RUN apt-get update && apt upgrade -y && \
  DEBIAN_FRONTEND=noninteractive apt-get install -yq --no-install-recommends \
    make curl git zip unzip wget dnsutils g++ gcc-aarch64-linux-gnu                 \
  && rm -rf /var/lib/apt/lists/*

RUN go version
ENV GO111MODULE=on
RUN go install github.com/go-delve/delve/cmd/dlv@latest
RUN go install github.com/cosmtrek/air@latest

WORKDIR /go/src/github.com/bloxapp/ssv/
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

CMD air

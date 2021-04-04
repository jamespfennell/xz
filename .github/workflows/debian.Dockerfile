# This file exists to have an easy way to do CI, not to actually
# distribute the resulting Docker image. Building the Docker image
# succesfully implies the code compiles and all the tests pass.
FROM golang:1.16 AS builder

WORKDIR /xz

# We install xz-utils to get the xz command line tool, which is used
# in tests.
# The package liblzma-dev is used to get the C library that is linked
# into the Go package.
RUN apt-get update && apt-get install xz-utils liblzma-dev

COPY ../../go.mod ./
COPY ../.. ./

RUN go test ./...
RUN go build cmd/goxz.go

FROM buildpack-deps:buster-curl

COPY --from=builder /xz/goxz /usr/bin

ENTRYPOINT ["goxz"]

# TODO:
#  - CentOS
#  - MacOS
#  - Alpine

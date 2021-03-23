# Copyright 2020 WhatsNew Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

FROM golang:1.15.7-alpine3.13 as Builder

# Build statically linked Go binaries without CGO.
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go mod download
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -v -x -a -o whatsnew cmd/whatsnew.go

# The definition of the runtime container now follows.
FROM alpine:3.13.2

RUN apk add --no-cache ca-certificates

# Copy the compiled program from the Builder container.
COPY --from=Builder /build/whatsnew whatsnew

# Setup the environment
ENV GOOS linux
ENV GOARCH amd64

# Run WhatsNew
ENTRYPOINT ["/whatsnew"]
CMD ["help"]

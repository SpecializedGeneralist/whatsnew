# Copyright 2020 WhatsNew Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

FROM golang:1.15.0-alpine3.12 as Builder

# Build statically linked Go binaries without CGO.
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go mod download
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-extldflags=-static" -v -x -a -o whatsnew cmd/whatsnew.go

# The definition of the runtime container now follows.
FROM alpine:3.12.0

RUN mkdir -p /app

# Copy the compiled program from the Builder container.
COPY --from=Builder /build/whatsnew /app/whatsnew

# Create configuration folder
RUN mkdir -p /app/config

# Setup the environment
ENV GOOS linux
ENV GOARCH amd64

# Run WhatsNew
ENTRYPOINT ["/app/whatsnew"]
CMD ["help"]
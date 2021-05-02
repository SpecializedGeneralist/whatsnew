# Copyright 2020-2021 WhatsNew Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

FROM golang:1.16.3-alpine3.13 as Builder

WORKDIR /go/src/whatsnew
COPY . .

RUN go mod download \
    && CGO_ENABLED=0 go build \
        -ldflags="-extldflags=-static" \
        -o /go/bin/whatsnew \
        cmd/whatsnew.go

FROM alpine:3.13.5

RUN apk add --no-cache ca-certificates

COPY --from=Builder /go/bin/whatsnew /bin/whatsnew

ENTRYPOINT ["/bin/whatsnew"]
CMD ["help"]

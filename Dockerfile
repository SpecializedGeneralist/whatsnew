# Copyright 2021 SpecializedGeneralist. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

FROM golang:1.17.1-alpine3.14 as Builder

WORKDIR /go/src/whatsnew
COPY . .

RUN go mod download \
    && go build \
        -ldflags="-extldflags=-static" \
        -o /go/bin/whatsnew \
        whatsnew.go

FROM alpine:3.14.2

RUN apk add --no-cache ca-certificates

COPY --from=Builder /go/bin/whatsnew /bin/whatsnew

ENTRYPOINT ["/bin/whatsnew"]

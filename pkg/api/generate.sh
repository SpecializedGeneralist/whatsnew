#!/usr/bin/env sh

# Copyright 2021 WhatsNew Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

set -ex

# Generate api.proto
gnostic \
  --grpc-out=. \
  api.yaml

sed \
  -i "2i option go_package = \"github.com/SpecializedGeneralist/whatsnew/pkg/api\";" \
  api.proto

# Generate api.pb.go
protoc \
  --go_out=. \
  --go_opt='paths=source_relative' \
  api.proto

# Generate api_grpc.pb.go
protoc \
  --go-grpc_out=. \
  --go-grpc_opt='paths=source_relative' \
  api.proto

# Generate api_descriptor.pb
protoc \
  --proto_path=. \
  --include_imports \
  --include_source_info \
  --descriptor_set_out=api_descriptor.pb \
  api.proto

# Generate api.pb.gw.go
protoc \
  --grpc-gateway_out=. \
  --grpc-gateway_opt='logtostderr=true' \
  --grpc-gateway_opt='paths=source_relative' \
  api.proto

// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package grpcconn provides utilities for handling gRPC client connections.
package grpcconn

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"time"
)

// Dial creates a client connection to the configured target, also respecting
// the given TLS configuration.
//
// This function blocks until the underlying connection is up, within a
// timeout of 30 seconds.
func Dial(ctx context.Context, conf config.GRPCServer) (*grpc.ClientConn, error) {
	ctxTO, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	opts := []grpc.DialOption{grpc.WithBlock()}
	if conf.TLSEnabled {
		creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.DialContext(ctxTO, conf.Target, opts...)
	if err != nil {
		return nil, fmt.Errorf("error dialing gRPC %+v: %w", conf, err)
	}
	return conn, nil
}

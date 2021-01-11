// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package grpcutils

import (
	"context"
	"crypto/tls"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// OpenConnection returns a new grpc.ClientConn object. It blocks until
// a connection is made or the process timed out.
func OpenConnection(address string, tlsDisable bool) (*grpc.ClientConn, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if tlsDisable {
		conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			return nil, err
		}
		return conn, nil
	}

	creds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})
	conn, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(creds), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

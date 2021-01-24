// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package labse

import (
	"context"
	"github.com/nlpodyssey/spago/pkg/mat32"

	spagogrpcapi "github.com/nlpodyssey/spago/pkg/nlp/transformers/bert/grpcapi"
	"google.golang.org/grpc"
)

// Gateway provides access to the LaBSE API (spaGO).
type Gateway struct {
	connection *grpc.ClientConn
}

// New returns Gateway objects for accessing the LaBSE API.
func New(connection *grpc.ClientConn) *Gateway {
	return &Gateway{
		connection: connection,
	}
}

// Vectorize returns a dense representation of given text using Language-Agnostic BERT Sentence Embedding (LaBSE).
func (s *Gateway) Vectorize(text string) ([]float32, error) {
	client := spagogrpcapi.NewBERTClient(s.connection)
	encoding, err := client.Encode(context.Background(), &spagogrpcapi.EncodeRequest{
		Text: text,
	})
	if err != nil {
		return nil, err
	}
	return normalize(encoding.Vector), nil
}

func (s *Gateway) Close() error {
	return s.connection.Close()
}

func normalize(xs []float32) []float32 {
	return mat32.NewVecDense(xs).Normalize2().Data()
}

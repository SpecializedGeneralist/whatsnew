// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zeroshotclassification

import (
	"context"
	spagogrpcapi "github.com/nlpodyssey/spago/pkg/nlp/transformers/bart/bartserver/grpcapi"
	"google.golang.org/grpc"
)

// Gateway provides access to spaGO classification API.
type Gateway struct {
	connection *grpc.ClientConn
}

// NewGateway returns a new Gateway.
func NewGateway(connection *grpc.ClientConn) *Gateway {
	return &Gateway{connection: connection}
}

// ClassifyNLI performs a zero-shot classification against spaGO API.
func (s *Gateway) ClassifyNLI(text, hypothesisTemplate string, possibleLabels []string, multiClass bool) (*ZeroShotClassification, error) {
	client := spagogrpcapi.NewBARTClient(s.connection)
	reply, err := client.ClassifyNLI(context.Background(), &spagogrpcapi.ClassifyNLIRequest{
		Text:               text,
		HypothesisTemplate: hypothesisTemplate,
		PossibleLabels:     possibleLabels,
		MultiClass:         multiClass,
	})
	if err != nil {
		return nil, err
	}

	dist := make([]ClassConfidencePair, len(reply.Distribution))
	for i, p := range reply.Distribution {
		dist[i] = ClassConfidencePair{
			Class:      p.Class,
			Confidence: float32(p.Confidence),
		}
	}

	return &ZeroShotClassification{
		Class:        reply.Class,
		Confidence:   float32(reply.Confidence),
		Distribution: dist,
		Took:         int(reply.Took),
	}, nil
}

// Close closes the connection.
func (s *Gateway) Close() error {
	return s.connection.Close()
}

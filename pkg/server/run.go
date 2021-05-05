// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/api"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strings"
)

// Run runs the server according to its configuration.
func (s *Server) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	grpcServer := grpc.NewServer()
	api.RegisterApiServer(grpcServer, s)

	gwmux := runtime.NewServeMux()
	err := api.RegisterApiHandlerServer(ctx, gwmux, s)
	if err != nil {
		return fmt.Errorf("failed to register service handler: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)

	listener, err := net.Listen("tcp", s.config.Address)
	if err != nil {
		return fmt.Errorf("TCP listen error: %w", err)
	}

	handler := handlerFunc(grpcServer, mux)

	if s.config.TLSEnabled {
		return s.serveTLS(listener, handler)
	}
	return s.serveInsecure(listener, handler)
}

func (s *Server) serveInsecure(listener net.Listener, handler http.Handler) error {
	h2s := &http2.Server{}
	h1s := &http.Server{
		Handler: h2c.NewHandler(handler, h2s),
	}

	s.logger.Info().Msgf("Serving on %s (insecure)", s.config.Address)
	err := h1s.Serve(listener)
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) serveTLS(listener net.Listener, handler http.Handler) error {
	tlsCert, err := tls.LoadX509KeyPair(s.config.TLSCert, s.config.TLSKey)
	if err != nil {
		return fmt.Errorf("error loading TLS certificate: %w", err)
	}

	hs := &http.Server{
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsCert},
			NextProtos:   []string{"h2"},
		},
	}

	s.logger.Info().Msgf("Serving on %s (TLS)", s.config.Address)
	err = hs.Serve(tls.NewListener(listener, hs.TLSConfig))
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func handlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isGRPCRequest(r) {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}

func isGRPCRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc")
}

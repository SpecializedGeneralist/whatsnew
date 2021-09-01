// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/server/whatsnew"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"strings"
)

// Run runs the server according to its configuration.
func (s *Server) Run(ctx context.Context) error {
	grpcServer := grpc.NewServer()
	whatsnew.RegisterWhatsnewServer(grpcServer, s)

	gwMux := runtime.NewServeMux()
	err := whatsnew.RegisterWhatsnewHandlerServer(ctx, gwMux, s)
	if err != nil {
		return fmt.Errorf("failed to register service handler: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwMux)

	listener, err := net.Listen("tcp", s.conf.Address)
	if err != nil {
		return fmt.Errorf("TCP listen error: %w", err)
	}

	handler := cors.New(s.corsOptions()).Handler(mux)
	handler = handlerFunc(grpcServer, handler)

	if s.conf.TLSEnabled {
		return s.serveTLS(ctx, listener, handler)
	}
	return s.serveInsecure(ctx, listener, handler)
}

func (s *Server) serveInsecure(ctx context.Context, listener net.Listener, handler http.Handler) error {
	h2s := &http2.Server{}
	h1s := &http.Server{
		Handler: h2c.NewHandler(handler, h2s),
	}

	s.log.Info().Msgf("Serving on %s (insecure)", s.conf.Address)

	go s.shutDownServerWhenContextIsDone(ctx, h1s)

	err := h1s.Serve(listener)
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) serveTLS(ctx context.Context, listener net.Listener, handler http.Handler) error {
	tlsCert, err := tls.LoadX509KeyPair(s.conf.TLSCert, s.conf.TLSKey)
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

	s.log.Info().Msgf("Serving on %s (TLS)", s.conf.Address)

	go s.shutDownServerWhenContextIsDone(ctx, hs)

	err = hs.Serve(tls.NewListener(listener, hs.TLSConfig))
	if err != nil {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

func (s *Server) shutDownServerWhenContextIsDone(ctx context.Context, hs *http.Server) {
	<-ctx.Done()
	s.log.Info().Msg("context done, shutting down server")
	err := hs.Shutdown(context.Background())
	if err != nil {
		s.log.Err(err).Msg("server shutdown error")
	}
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

func (s *Server) corsOptions() cors.Options {
	return cors.Options{
		AllowedOrigins: s.conf.AllowedOrigins,
	}
}

func isGRPCRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc")
}

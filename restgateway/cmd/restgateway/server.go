package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
)

// Server is the root object of server
type Server struct {
	server          http.Server
	addr            string

	listener net.Listener
}

// New returns a new server configured with the provided options
func New(address string, handler http.Handler) (*Server, error) {
	srvr := &Server{server: http.Server{Handler: handler}}

	var err error
	if srvr.listener, err = net.Listen("tcp", address); err != nil {
		return nil, fmt.Errorf("net.Listen: %w", err)
	}

	return srvr, nil
}

// Run blocks and lasts either as long as the context or until an error is encountered.
func (s *Server) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	errs := make(chan error, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		errs <- s.server.Serve(s.listener) // chan is buffered so putting on it never blocks
	}()

	fmt.Println(fmt.Sprintf("listener started on \"%s\"", s.listener.Addr().String()))

	select {

	case err := <-errs:
		return fmt.Errorf("server.Server: %w", err)

	case <-ctx.Done():
		// Shutdown means that the s.ListenAndServe call immediately return with ErrServerClosed; we ignore that error,
		// and the goroutine can immediately return because of the buffered channel
		if err := s.server.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("server.Shutdown: %w", err)
		}
		return nil
	}
}

// Close closes any initialized resources.
func (s *Server) Close() error {
	return s.listener.Close()
}
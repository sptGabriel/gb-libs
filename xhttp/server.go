package xhttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type Config struct {
	Addr            string        `env:"HTTP_ADDR" envDefault:":8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"2m"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

type Server struct {
	cfg Config
	srv *http.Server
}

// NewServer wraps the given handler with the default middleware chain
// (request data, panic recovery and access logging).
func NewServer(cfg Config, handler http.Handler) *Server {
	return &Server{
		cfg: cfg,
		srv: &http.Server{
			Addr:         cfg.Addr,
			Handler:      RequestData(Recover(AccessLog(handler))),
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
	}
}

// Run blocks until ctx is canceled, then shuts down gracefully. Request
// contexts inherit values from ctx (logger, tracer) but not its cancellation,
// so in-flight requests can finish during the shutdown grace period.
func (s *Server) Run(ctx context.Context) error {
	s.srv.BaseContext = func(net.Listener) context.Context {
		return context.WithoutCancel(ctx)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()

	return s.srv.Shutdown(shutdownCtx)
}

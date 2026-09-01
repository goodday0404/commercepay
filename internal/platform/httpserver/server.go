package httpserver

import (
	"net/http"
	"time"
)

type ServerOptions struct {
	Address           string
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

func NewServer(opts ServerOptions, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              opts.Address,
		ReadHeaderTimeout: opts.ReadHeaderTimeout,
		ReadTimeout:       opts.ReadTimeout,
		WriteTimeout:      opts.WriteTimeout,
		IdleTimeout:       opts.IdleTimeout,
		Handler:           handler,
	}
}

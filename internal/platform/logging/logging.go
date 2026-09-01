package logging

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Options struct {
	Level       string
	Format      string
	AddSource   bool
	ServiceName string
	Environment string
}

func New(opts Options) (*slog.Logger, error) {
	var logLevel slog.Level

	if err := logLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		return nil, fmt.Errorf(
			"invalid log level %q: %w",
			opts.Level,
			err,
		)
	}

	handlerOptions := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: opts.AddSource,
	}

	var handler slog.Handler

	switch strings.ToLower(opts.Format) {
	case "json":
		handler = slog.NewJSONHandler(
			os.Stdout,
			handlerOptions,
		)

	case "text":
		handler = slog.NewTextHandler(
			os.Stdout,
			handlerOptions,
		)

	default:
		return nil, fmt.Errorf(
			"invalid log format %q",
			opts.Format,
		)
	}

	logger := slog.New(handler).With(
		"service", opts.ServiceName,
		"environment", opts.Environment,
	)

	return logger, nil
}

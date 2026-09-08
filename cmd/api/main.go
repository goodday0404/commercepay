package main

import (
	"fmt"
	"os"

	"github.com/goodday0404/commercepay/internal/platform/config"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "CommercePay failed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	deps, err := buildDependencies(cfg)
	if err != nil {
		return fmt.Errorf("build dependencies: %w", err)
	}
	defer deps.Close()

	handlers := buildHandlers(deps)
	registerRoutes(deps.Router, handlers)

	server := buildHTTPServer(cfg, deps)

	err = runServer(server, deps.Logger)

	return err
}

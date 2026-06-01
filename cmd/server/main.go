package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cristian-scherer/memoos/internal/app"
	"github.com/cristian-scherer/memoos/internal/mcp"
)

func main() {
	application, err := app.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize app: %v\n", err)
		os.Exit(1)
	}
	defer application.Stop()

	if err := application.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start app: %v\n", err)
		os.Exit(1)
	}

	mcpServer := mcp.NewServer(application.Logger)
	toolHandler := mcp.NewToolHandler(application.MemService)
	toolHandler.RegisterTools(mcpServer)

	ctx := application.Context()
	if err := mcpServer.Start(ctx); err != nil {
		application.Logger.Errorf("Failed to start MCP server: %v", err)
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	application.Logger.Info("Shutting down...")
}

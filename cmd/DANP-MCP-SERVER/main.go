package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IceFireLabs/DANP-Engine/core/mcp"
)

func main() {
	// Create a context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	// Create and start the server
	server, err := mcp.NewServer("config/mcp_manifest.yaml")
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	log.Printf("MCP server created successfully with config: %s", "config/mcp_manifest.yaml")
	log.Printf("Server info: %+v", server)

	// Start the server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("MCP server failed: %v", err)
			cancel()
		}
	}()

	log.Println("MCP server started and running")
	
	// Wait for context cancellation
	<-ctx.Done()
	log.Println("Shutting down MCP server")
	
	// Gracefully stop the server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	
	if err := server.Stop(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}
}

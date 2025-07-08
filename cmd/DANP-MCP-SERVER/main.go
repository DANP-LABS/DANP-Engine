package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DANP-LABS/DANP-Engine/core/mcp"
)

// setupWallet handles the creation or loading of the wallet.
func setupWallet(walletPath string) (*mcp.Wallet, error) {
	log.Println("Initiating wallet setup...")

	walletPassword := os.Getenv("WALLET_PASSWORD")
	if walletPassword == "" {
		return nil, fmt.Errorf("WALLET_PASSWORD environment variable not set. Please set it to secure your wallet")
	}
	log.Println("WALLET_PASSWORD environment variable found.")

	log.Printf("Attempting to load or create wallet from %s...", walletPath)
	wallet, err := mcp.GetOrCreateWallet(walletPath, walletPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create wallet: %w", err)
	}

	log.Printf("Wallet loaded successfully. Address: %s", wallet.Address.Hex())
	log.Println("Wallet setup complete.")
	return wallet, nil
}

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

	// Wallet setup
	wallet, err := setupWallet("config/wallet.json")
	if err != nil {
		log.Fatalf("Wallet setup failed: %v", err)
	}
	log.Printf("Server is using wallet address: %s", wallet.Address.Hex())

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


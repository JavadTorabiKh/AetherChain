package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/javadtorabikh/Aetherchain/config"
	"github.com/javadtorabikh/Aetherchain/blockchain"
	"github.com/javadtorabikh/Aetherchain/network"
	"github.com/javadtorabikh/Aetherchain/api"
)

// @title AetherChain Full Node
// @version 1.0
// @description A complete blockchain full node implementation in Go
// @contact.name AetherChain Team
// @contact.url https://github.com/your-username/aetherchain
func main() {
	fmt.Println(`
    ___       __  __           _    _           _       
   /   | ____/ /_/ /_  _______| |  / /__  _____(_)___ _ 
  / /| |/ __/ __/ / / / / ___/ | / / _ \/ ___/ / __  /
 / ___ / /_/ /_/ / /_/ (__  )| |/ /  __/ /  / / /_/ / 
/_/  |_\__/\__/_/\__,_/____/ |___/\___/_/  /_/\__,_/  
                                                       
🚀 Starting AetherChain Full Node...
	`)

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize blockchain
	bc := blockchain.NewBlockchain(cfg.Difficulty, cfg.BlockReward)
	fmt.Printf("📦 Blockchain initialized with genesis block\n")

	// Initialize network node
	node := network.NewNode(cfg, bc)
	
	// Start network services
	if err := node.Start(); err != nil {
		log.Fatalf("Failed to start network node: %v", err)
	}
	fmt.Printf("🌐 Network node started on %s:%d\n", cfg.Host, cfg.Port)

	// Start API server if enabled
	if cfg.APIEnabled {
		apiServer := api.NewServer(cfg, bc, node)
		go func() {
			if err := apiServer.Start(); err != nil {
				log.Printf("API server error: %v", err)
			}
		}()
		fmt.Printf("🔗 API server started on %s:%d\n", cfg.APIHost, cfg.APIPort)
		fmt.Printf("📚 API Documentation: http://%s:%d/docs\n", cfg.APIHost, cfg.APIPort)
	}

	// Display node information
	fmt.Printf("\n")
	fmt.Printf("📍 Node ID: %s\n", cfg.NodeID)
	fmt.Printf("🌍 Environment: %s\n", cfg.Environment)
	fmt.Printf("⛓️  Chain Height: %d\n", len(bc.Chain))
	fmt.Printf("🎯 Difficulty: %d\n", cfg.Difficulty)
	fmt.Printf("💰 Block Reward: %.2f\n", cfg.BlockReward)
	fmt.Printf("\n")

	// Wait for interrupt signal to gracefully shutdown
	waitForShutdown(node)

	fmt.Println("👋 AetherChain node stopped gracefully")
}

// waitForShutdown handles graceful shutdown on interrupt signals
func waitForShutdown(node *network.Node) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigCh
	fmt.Println("\n🛑 Received shutdown signal...")
	
	// Graceful shutdown
	node.Stop()
}
package api

import (

    "time" 
)

import "github.com/gin-gonic/gin"

// setupRoutes configures all API routes with middleware
func (s *Server) setupRoutes() {
	// Apply global middleware
	s.router.Use(Logger())
	s.router.Use(CORS())
	s.router.Use(Recovery())
	s.router.Use(RateLimiter())

	// Health check endpoint
	s.router.GET("/health", s.healthCheck)

	// API documentation
	s.router.GET("/", s.getDocumentation)
	s.router.GET("/docs", s.getDocumentation)

	// API v1 routes
	apiV1 := s.router.Group("/api/v1")
	{
		// Apply authentication to all API v1 routes (optional)
		// apiV1.Use(AuthMiddleware())
		
		// Blockchain endpoints
		blockchain := apiV1.Group("/blockchain")
		{
			blockchain.GET("/info", s.getBlockchainInfo)
			blockchain.GET("/blocks", s.getBlocks)
			blockchain.GET("/blocks/:height", s.getBlockByHeight)
			blockchain.GET("/blocks/hash/:hash", s.getBlockByHash)
			blockchain.GET("/transactions/pending", s.getPendingTransactions)
			blockchain.GET("/transactions/:hash", s.getTransaction)
			blockchain.POST("/transactions", s.createTransaction)
			blockchain.GET("/balance/:address", s.getBalance)
			blockchain.GET("/validity", s.checkChainValidity)
		}

		// Mining endpoints
		mining := apiV1.Group("/mining")
		{
			mining.GET("/mine", s.mineBlock)
			mining.GET("/status", s.getMiningStatus)
			mining.GET("/reward", s.getBlockReward)
		}

		// Network endpoints
		network := apiV1.Group("/network")
		{
			network.GET("/info", s.getNetworkInfo)
			network.GET("/peers", s.getPeers)
			network.POST("/peers", s.addPeer)
			network.GET("/discovery", s.getDiscoveredPeers)
			network.GET("/stats", s.getNetworkStats)
		}

		// Node endpoints
		node := apiV1.Group("/node")
		{
			node.GET("/status", s.getNodeStatus)
			node.GET("/version", s.getVersion)
			node.GET("/config", s.getNodeConfig)
			node.POST("/restart", s.restartNode)
		}

		// Wallet endpoints (basic)
		wallet := apiV1.Group("/wallet")
		{
			wallet.POST("/create", s.createWallet)
			wallet.GET("/addresses", s.getAddresses)
		}
	}
}

// healthCheck returns the health status of the node
func (s *Server) healthCheck(c *gin.Context) {
	// Check if blockchain is valid
	isValid := s.blockchain.IsChainValid()
	
	// Check if node is running
	nodeRunning := true // This would check actual node status
	
	healthStatus := "healthy"
	if !isValid || !nodeRunning {
		healthStatus = "unhealthy"
	}

	c.JSON(200, gin.H{
		"status":    healthStatus,
		"timestamp": time.Now().Unix(),
		"checks": gin.H{
			"blockchain_valid": isValid,
			"node_running":     nodeRunning,
			"peers_connected":  s.node.GetPeerCount() > 0,
		},
		"version": s.config.Version,
	})
}

// getPendingTransactions returns pending transactions from the pool
func (s *Server) getPendingTransactions(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"transactions": s.blockchain.TransactionPool,
			"count":        len(s.blockchain.TransactionPool),
		},
	})
}

// checkChainValidity checks if the blockchain is valid
func (s *Server) checkChainValidity(c *gin.Context) {
	isValid := s.blockchain.IsChainValid()
	
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"is_valid": isValid,
			"message":  "Blockchain validation completed",
		},
	})
}

// getBlockReward returns current block reward
func (s *Server) getBlockReward(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"block_reward": s.blockchain.BlockReward,
		},
	})
}

// getDiscoveredPeers returns discovered peers
func (s *Server) getDiscoveredPeers(c *gin.Context) {
	// This would return peers discovered through peer discovery
	// For now, return empty list
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"discovered_peers": []string{},
			"count": 0,
		},
	})
}

// getNetworkStats returns network statistics
func (s *Server) getNetworkStats(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"total_peers":      s.node.GetPeerCount(),
			"bytes_sent":       0, // Would track actual network usage
			"bytes_received":   0,
			"connections":      s.node.GetPeerCount(),
			"uptime":           "0", // Would track node uptime
		},
	})
}

// getNodeConfig returns node configuration (without sensitive info)
func (s *Server) getNodeConfig(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"node_id":      s.config.NodeID,
			"version":      s.config.Version,
			"environment":  s.config.Environment,
			"api_enabled":  s.config.APIEnabled,
			"api_host":     s.config.APIHost,
			"api_port":     s.config.APIPort,
			"difficulty":   s.config.Difficulty,
			"block_reward": s.config.BlockReward,
		},
	})
}

// restartNode restarts the node (placeholder)
func (s *Server) restartNode(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Node restart initiated",
			"note":    "This is a placeholder - actual restart would be implemented",
		},
	})
}

// createWallet creates a new wallet (placeholder)
func (s *Server) createWallet(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Wallet creation endpoint",
			"note":    "This would generate new cryptographic key pairs",
		},
	})
}

// getAddresses returns wallet addresses (placeholder)
func (s *Server) getAddresses(c *gin.Context) {
	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"addresses": []string{},
			"note":     "This would return addresses from the wallet",
		},
	})
}
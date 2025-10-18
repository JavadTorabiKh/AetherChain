package api

import (
	"fmt"
	"net/http"
	"strconv"

	"your-username/aetherchain/config"
	"your-username/aetherchain/blockchain"
	"your-username/aetherchain/network"

	"github.com/gin-gonic/gin"
)

// Server represents the API server
type Server struct {
	config    *config.Config
	blockchain *blockchain.Blockchain
	node      *network.Node
	router    *gin.Engine
}

// NewServer creates a new API server instance
func NewServer(cfg *config.Config, bc *blockchain.Blockchain, node *network.Node) *Server {
	server := &Server{
		config:    cfg,
		blockchain: bc,
		node:      node,
		router:    gin.Default(),
	}

	server.setupRoutes()
	return server
}

// Start begins the API server
func (s *Server) Start() error {
	address := fmt.Sprintf("%s:%d", s.config.APIHost, s.config.APIPort)
	fmt.Printf("🌐 API server starting on %s\n", address)
	
	return s.router.Run(address)
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// API documentation
	s.router.GET("/", s.getDocumentation)
	s.router.GET("/docs", s.getDocumentation)

	// Blockchain endpoints
	blockchainGroup := s.router.Group("/api/v1/blockchain")
	{
		blockchainGroup.GET("/info", s.getBlockchainInfo)
		blockchainGroup.GET("/blocks", s.getBlocks)
		blockchainGroup.GET("/blocks/:height", s.getBlockByHeight)
		blockchainGroup.GET("/blocks/hash/:hash", s.getBlockByHash)
		blockchainGroup.GET("/transactions/:hash", s.getTransaction)
		blockchainGroup.POST("/transactions", s.createTransaction)
		blockchainGroup.GET("/balance/:address", s.getBalance)
	}

	// Mining endpoints
	miningGroup := s.router.Group("/api/v1/mining")
	{
		miningGroup.GET("/mine", s.mineBlock)
		miningGroup.GET("/status", s.getMiningStatus)
	}

	// Network endpoints
	networkGroup := s.router.Group("/api/v1/network")
	{
		networkGroup.GET("/info", s.getNetworkInfo)
		networkGroup.GET("/peers", s.getPeers)
		networkGroup.POST("/peers", s.addPeer)
	}

	// Node endpoints
	nodeGroup := s.router.Group("/api/v1/node")
	{
		nodeGroup.GET("/status", s.getNodeStatus)
		nodeGroup.GET("/version", s.getVersion)
	}
}

// getDocumentation returns API documentation
func (s *Server) getDocumentation(c *gin.Context) {
	docs := gin.H{
		"name":        "AetherChain API",
		"version":     s.config.Version,
		"description": "Complete blockchain full node API",
		"endpoints": gin.H{
			"blockchain": gin.H{
				"GET /api/v1/blockchain/info":           "Get blockchain information",
				"GET /api/v1/blockchain/blocks":         "Get all blocks",
				"GET /api/v1/blockchain/blocks/:height": "Get block by height",
				"GET /api/v1/blockchain/balance/:address": "Get address balance",
				"POST /api/v1/blockchain/transactions":  "Create new transaction",
			},
			"mining": gin.H{
				"GET /api/v1/mining/mine":   "Mine a new block",
				"GET /api/v1/mining/status": "Get mining status",
			},
			"network": gin.H{
				"GET /api/v1/network/info":  "Get network information",
				"GET /api/v1/network/peers": "Get connected peers",
				"POST /api/v1/network/peers": "Add new peer",
			},
			"node": gin.H{
				"GET /api/v1/node/status":  "Get node status",
				"GET /api/v1/node/version": "Get node version",
			},
		},
	}

	c.JSON(http.StatusOK, docs)
}

// getBlockchainInfo returns blockchain information
func (s *Server) getBlockchainInfo(c *gin.Context) {
	info := s.blockchain.GetChainInfo()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// getBlocks returns all blocks in the blockchain
func (s *Server) getBlocks(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"blocks": s.blockchain.Chain,
			"count":  len(s.blockchain.Chain),
		},
	})
}

// getBlockByHeight returns a specific block by height
func (s *Server) getBlockByHeight(c *gin.Context) {
	heightStr := c.Param("height")
	height, err := strconv.Atoi(heightStr)
	if err != nil || height < 0 || height >= len(s.blockchain.Chain) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid block height",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    s.blockchain.Chain[height],
	})
}

// getBlockByHash returns a block by its hash
func (s *Server) getBlockByHash(c *gin.Context) {
	hash := c.Param("hash")
	
	for _, block := range s.blockchain.Chain {
		if block.Hash == hash {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    block,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Block not found",
	})
}

// getTransaction returns a transaction by hash
func (s *Server) getTransaction(c *gin.Context) {
	// This would search through all blocks for the transaction
	// Simplified implementation for now
	c.JSON(http.StatusNotImplemented, gin.H{
		"success": false,
		"error":   "Not implemented yet",
	})
}

// createTransaction creates a new transaction
func (s *Server) createTransaction(c *gin.Context) {
	var txRequest struct {
		From   string  `json:"from" binding:"required"`
		To     string  `json:"to" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
		Fee    float64 `json:"fee"`
	}

	if err := c.ShouldBindJSON(&txRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Create transaction
	tx := blockchain.NewTransaction(
		txRequest.From,
		txRequest.To,
		txRequest.Amount,
		txRequest.Fee,
		time.Now().UnixNano(), // Using timestamp as nonce for simplicity
	)

	// Add to blockchain
	if err := s.blockchain.AddTransaction(tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"transaction": tx,
			"message":    "Transaction added to pool",
		},
	})
}

// getBalance returns the balance of an address
func (s *Server) getBalance(c *gin.Context) {
	address := c.Param("address")
	balance := s.blockchain.GetBalance(address)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"address": address,
			"balance": balance,
		},
	})
}

// mineBlock mines a new block
func (s *Server) mineBlock(c *gin.Context) {
	minerAddress := c.DefaultQuery("miner", "default_miner")
	
	block, err := s.blockchain.CreateNewBlock(minerAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"block":   block,
			"message": "Block mined successfully",
		},
	})
}

// getMiningStatus returns mining status
func (s *Server) getMiningStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"mining":       false, // This would track actual mining status
			"difficulty":   s.blockchain.Difficulty,
			"block_reward": s.blockchain.BlockReward,
		},
	})
}

// getNetworkInfo returns network information
func (s *Server) getNetworkInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"node_id":      s.config.NodeID,
			"peers_count":  s.node.GetPeerCount(),
			"host":         s.config.Host,
			"port":         s.config.Port,
			"environment":  s.config.Environment,
		},
	})
}

// getPeers returns connected peers
func (s *Server) getPeers(c *gin.Context) {
	peers := s.node.GetPeerList()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"peers": peers,
			"count": len(peers),
		},
	})
}

// addPeer adds a new peer
func (s *Server) addPeer(c *gin.Context) {
	var peerRequest struct {
		Address string `json:"address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&peerRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	go s.node.connectToNode(peerRequest.Address)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Connecting to peer",
			"address": peerRequest.Address,
		},
	})
}

// getNodeStatus returns node status
func (s *Server) getNodeStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status":      "running",
			"uptime":      "0", // This would track actual uptime
			"block_height": len(s.blockchain.Chain),
			"sync_status": "synced",
		},
	})
}

// getVersion returns node version
func (s *Server) getVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"version":     s.config.Version,
			"name":       "AetherChain",
			"network":    s.config.Environment,
		},
	})
}
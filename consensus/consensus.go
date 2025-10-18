package consensus

import (
	"fmt"
	"sync"
	"time"

	"Aetherchain/blockchain"
	"Aetherchain/network"
)

// Consensus implements the consensus mechanism for AetherChain
type Consensus struct {
	blockchain *blockchain.Blockchain
	node       *network.Node
	isMining   bool
	miningStop chan bool
	mutex      sync.RWMutex
}

// NewConsensus creates a new consensus instance
func NewConsensus(bc *blockchain.Blockchain, node *network.Node) *Consensus {
	return &Consensus{
		blockchain: bc,
		node:       node,
		miningStop: make(chan bool),
	}
}

// StartMining begins the mining process
func (c *Consensus) StartMining(minerAddress string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isMining {
		return fmt.Errorf("mining is already in progress")
	}

	c.isMining = true
	fmt.Printf("⛏️ Starting mining with address: %s\n", minerAddress)

	// Start mining in a separate goroutine
	go c.miningLoop(minerAddress)

	return nil
}

// StopMining stops the mining process
func (c *Consensus) StopMining() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.isMining {
		c.miningStop <- true
		c.isMining = false
		fmt.Println("⛏️ Mining stopped")
	}
}

// miningLoop is the main mining loop
func (c *Consensus) miningLoop(minerAddress string) {
	miningTicker := time.NewTicker(10 * time.Second) // Check for new transactions every 10 seconds
	defer miningTicker.Stop()

	for {
		select {
		case <-c.miningStop:
			return
		case <-miningTicker.C:
			// Only mine if there are pending transactions
			if len(c.blockchain.TransactionPool) > 0 {
				c.mineBlock(minerAddress)
			} else {
				fmt.Println("⏳ No transactions to mine, waiting...")
			}
		}
	}
}

// mineBlock attempts to mine a new block
func (c *Consensus) mineBlock(minerAddress string) {
	fmt.Printf("⛏️ Attempting to mine new block with %d pending transactions...\n", 
		len(c.blockchain.TransactionPool))

	// Create and mine new block
	block, err := c.blockchain.CreateNewBlock(minerAddress)
	if err != nil {
		fmt.Printf("❌ Mining failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully mined block %d\n", block.Index)
	fmt.Printf("📦 Block hash: %s\n", block.Hash)
	fmt.Printf("💰 Miner reward: %.2f\n", block.BlockReward)

	// Add block to blockchain
	if err := c.blockchain.AddBlock(block); err != nil {
		fmt.Printf("❌ Failed to add mined block: %v\n", err)
		return
	}

	// Broadcast new block to network
	c.broadcastNewBlock(block)
}

// broadcastNewBlock broadcasts a newly mined block to the network
func (c *Consensus) broadcastNewBlock(block *blockchain.Block) {
	// In a real implementation, this would use the network layer to broadcast
	// For now, we'll just log the action
	fmt.Printf("📢 Broadcasting new block %d to network\n", block.Index)
	
	// This would typically use: c.node.BroadcastNewBlock(block)
}

// IsMining returns whether the node is currently mining
func (c *Consensus) IsMining() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.isMining
}

// GetMiningStatus returns detailed mining status
func (c *Consensus) GetMiningStatus() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"is_mining":          c.isMining,
		"miner_address":      "default_miner", // This would track the actual miner
		"pending_transactions": len(c.blockchain.TransactionPool),
		"difficulty":         c.blockchain.Difficulty,
		"block_reward":       c.blockchain.BlockReward,
	}
}

// ValidateBlock validates a block against consensus rules
func (c *Consensus) ValidateBlock(block *blockchain.Block) bool {
	// Check basic block validity
	if !block.IsValid() {
		return false
	}

	// Check if block follows the chain
	lastBlock := c.blockchain.GetLastBlock()
	if block.Index != lastBlock.Index+1 {
		fmt.Printf("❌ Block index mismatch: expected %d, got %d\n", 
			lastBlock.Index+1, block.Index)
		return false
	}

	if block.PrevHash != lastBlock.Hash {
		fmt.Printf("❌ Block previous hash mismatch\n")
		return false
	}

	// Validate proof of work
	pow := blockchain.NewProofOfWork(block, c.blockchain.Difficulty)
	if !pow.Validate() {
		fmt.Printf("❌ Block proof of work invalid\n")
		return false
	}

	// Validate all transactions in the block
	for _, tx := range block.Transactions {
		if !tx.IsValid() {
			fmt.Printf("❌ Block contains invalid transaction: %s\n", tx.Hash)
			return false
		}
	}

	fmt.Printf("✅ Block %d validated successfully\n", block.Index)
	return true
}

// HandleReceivedBlock processes a block received from the network
func (c *Consensus) HandleReceivedBlock(block *blockchain.Block) {
	fmt.Printf("📦 Received block %d from network\n", block.Index)

	// Validate the block
	if !c.ValidateBlock(block) {
		fmt.Printf("❌ Received block %d failed validation\n", block.Index)
		return
	}

	// Add to blockchain
	if err := c.blockchain.AddBlock(block); err != nil {
		fmt.Printf("❌ Failed to add received block: %v\n", err)
		return
	}

	fmt.Printf("✅ Successfully added received block %d to chain\n", block.Index)

	// If we're mining, we might want to stop current mining attempt
	// since a new block was added to the chain
	if c.IsMining() {
		fmt.Println("⏸️ New block received, mining may need to restart...")
	}
}
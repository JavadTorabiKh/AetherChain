package blockchain

import (
	"fmt"
	"sync"
	"time"
)

// Blockchain represents the complete AetherChain blockchain
type Blockchain struct {
    Chain        []*Block          `json:"chain"`
    PendingTx    []*Transaction    `json:"pending_transactions"`
    Difficulty   int               `json:"difficulty"`
    BlockReward  float64           `json:"block_reward"`
    
    // State management
    Accounts     map[string]float64 `json:"accounts"` // Address -> Balance
    TransactionPool []*Transaction  `json:"transaction_pool"`
    
    // Concurrency control
    mutex sync.RWMutex
}

// NewBlockchain creates and initializes a new blockchain
func NewBlockchain(difficulty int, blockReward float64) *Blockchain {
    bc := &Blockchain{
        Difficulty:  difficulty,
        BlockReward: blockReward,
        Accounts:    make(map[string]float64),
    }
    
    // Create and add the genesis block
    bc.CreateGenesisBlock()
    
    return bc
}

// CreateGenesisBlock creates the first block in the blockchain
func (bc *Blockchain) CreateGenesisBlock() {
    bc.mutex.Lock()
    defer bc.mutex.Unlock()
    
    genesisTransactions := []*Transaction{
        {
            Version:   1,
            Hash:      "genesis_transaction",
            From:      "0",
            To:        "genesis_address",
            Amount:    1000000,
            Fee:       0,
            Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
            Status:    "confirmed",
        },
    }
    
    genesisBlock := NewBlock(0, genesisTransactions, "0", bc.Difficulty)
    genesisBlock.Hash = genesisBlock.CalculateHash()
    genesisBlock.Miner = "genesis_miner"
    
    bc.Chain = []*Block{genesisBlock}
    
    // Initialize genesis account
    bc.Accounts["genesis_address"] = 1000000
}

// AddBlock adds a new block to the blockchain after validation
func (bc *Blockchain) AddBlock(block *Block) error {
    bc.mutex.Lock()
    defer bc.mutex.Unlock()
    
    // Validate the block
    if !bc.IsValidBlock(block) {
        return fmt.Errorf("invalid block")
    }
    
    // Process transactions in the block
    for _, tx := range block.Transactions {
        bc.processTransaction(tx)
    }
    
    // Add miner reward
    bc.Accounts[block.Miner] += block.BlockReward
    
    // Add block to chain
    bc.Chain = append(bc.Chain, block)
    
    // Remove processed transactions from pool
    bc.removeProcessedTransactions(block.Transactions)
    
    return nil
}

// CreateNewBlock creates a new block with pending transactions
func (bc *Blockchain) CreateNewBlock(miner string) (*Block, error) {
    bc.mutex.RLock()
    defer bc.mutex.RUnlock()
    
    if len(bc.Chain) == 0 {
        return nil, fmt.Errorf("blockchain not initialized")
    }
    
    lastBlock := bc.Chain[len(bc.Chain)-1]
    
    // Get transactions from pool (limit block size)
    transactions := bc.getTransactionsForBlock()
    
    newBlock := NewBlock(len(bc.Chain), transactions, lastBlock.Hash, bc.Difficulty)
    newBlock.Miner = miner
    
    // Mine the block
    pow := NewProofOfWork(newBlock, bc.Difficulty)
    nonce, hash, err := pow.Mine()
    if err != nil {
        return nil, err
    }
    
    newBlock.Nonce = nonce
    newBlock.Hash = hash
    
    return newBlock, nil
}

// AddTransaction adds a new transaction to the pool
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
    bc.mutex.Lock()
    defer bc.mutex.Unlock()
    
    if !tx.IsValid() {
        return fmt.Errorf("invalid transaction")
    }
    
    // Check if sender has sufficient balance
    if bc.Accounts[tx.From] < tx.Amount+tx.Fee {
        return fmt.Errorf("insufficient balance")
    }
    
    bc.TransactionPool = append(bc.TransactionPool, tx)
    return nil
}

// IsValidBlock validates a block before adding to the chain
func (bc *Blockchain) IsValidBlock(block *Block) bool {
    if block == nil {
        return false
    }
    
    // Check block index
    if block.Index != len(bc.Chain) {
        return false
    }
    
    // Check previous hash
    lastBlock := bc.Chain[len(bc.Chain)-1]
    if block.PrevHash != lastBlock.Hash {
        return false
    }
    
    // Validate proof of work
    pow := NewProofOfWork(block, bc.Difficulty)
    if !pow.Validate() {
        return false
    }
    
    // Validate all transactions in the block
    for _, tx := range block.Transactions {
        if !tx.IsValid() {
            return false
        }
    }
    
    return true
}

// GetBalance returns the balance of an address
func (bc *Blockchain) GetBalance(address string) float64 {
    bc.mutex.RLock()
    defer bc.mutex.RUnlock()
    
    return bc.Accounts[address]
}

// GetLastBlock returns the most recent block in the chain
func (bc *Blockchain) GetLastBlock() *Block {
    bc.mutex.RLock()
    defer bc.mutex.RUnlock()
    
    if len(bc.Chain) == 0 {
        return nil
    }
    
    return bc.Chain[len(bc.Chain)-1]
}

// IsChainValid validates the entire blockchain
func (bc *Blockchain) IsChainValid() bool {
    bc.mutex.RLock()
    defer bc.mutex.RUnlock()
    
    for i := 1; i < len(bc.Chain); i++ {
        currentBlock := bc.Chain[i]
        previousBlock := bc.Chain[i-1]
        
        // Check block hash
        if currentBlock.Hash != currentBlock.CalculateHash() {
            return false
        }
        
        // Check chain linkage
        if currentBlock.PrevHash != previousBlock.Hash {
            return false
        }
        
        // Check proof of work
        pow := NewProofOfWork(currentBlock, bc.Difficulty)
        if !pow.Validate() {
            return false
        }
    }
    
    return true
}

// Helper functions
func (bc *Blockchain) processTransaction(tx *Transaction) {
    // Deduct from sender
    bc.Accounts[tx.From] -= tx.Amount + tx.Fee
    // Add to recipient
    bc.Accounts[tx.To] += tx.Amount
    // Miner gets the fee (will be added when block is processed)
}

func (bc *Blockchain) removeProcessedTransactions(processed []*Transaction) {
    var remaining []*Transaction
    processedMap := make(map[string]bool)
    
    for _, tx := range processed {
        processedMap[tx.Hash] = true
    }
    
    for _, tx := range bc.TransactionPool {
        if !processedMap[tx.Hash] {
            remaining = append(remaining, tx)
        }
    }
    
    bc.TransactionPool = remaining
}

func (bc *Blockchain) getTransactionsForBlock() []*Transaction {
    // Simple implementation: take first 100 transactions
    // In production, this would prioritize by fee
    maxTransactions := 100
    if len(bc.TransactionPool) < maxTransactions {
        return bc.TransactionPool
    }
    return bc.TransactionPool[:maxTransactions]
}

// GetChainInfo returns basic blockchain information
func (bc *Blockchain) GetChainInfo() map[string]interface{} {
    bc.mutex.RLock()
    defer bc.mutex.RUnlock()
    
    return map[string]interface{}{
        "height":          len(bc.Chain),
        "difficulty":      bc.Difficulty,
        "block_reward":    bc.BlockReward,
        "pending_txs":     len(bc.TransactionPool),
        "total_accounts":  len(bc.Accounts),
        "last_block_hash": bc.Chain[len(bc.Chain)-1].Hash,
    }
}
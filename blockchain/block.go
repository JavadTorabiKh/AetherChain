package blockchain

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "time"
)

// Block represents a single block in the AetherChain blockchain
type Block struct {
    // Header
    Version    int    `json:"version"`     // Block version for protocol upgrades
    Index      int    `json:"index"`       // Block height in the chain
    Timestamp  int64  `json:"timestamp"`   // Unix timestamp of block creation
    PrevHash   string `json:"prev_hash"`   // Hash of the previous block
    MerkleRoot string `json:"merkle_root"` // Merkle root of transactions
    
    // Body
    Transactions []*Transaction `json:"transactions"` // List of transactions
    Nonce        int64          `json:"nonce"`        // Proof-of-Work nonce
    Difficulty   int            `json:"difficulty"`   // Mining difficulty
    
    // Metadata
    Hash         string  `json:"hash"`          // Current block hash
    Miner        string  `json:"miner"`         // Miner's address
    BlockReward  float64 `json:"block_reward"`  // Reward for mining this block
}

// NewBlock creates a new block with the given parameters
func NewBlock(index int, transactions []*Transaction, prevHash string, difficulty int) *Block {
    block := &Block{
        Version:      1,
        Index:        index,
        Timestamp:    time.Now().Unix(),
        PrevHash:     prevHash,
        Transactions: transactions,
        Difficulty:   difficulty,
        BlockReward:  50.0, // Base block reward
    }
    
    // Calculate Merkle root from transactions
    block.MerkleRoot = block.CalculateMerkleRoot()
    
    return block
}

// CalculateHash computes and returns the SHA-256 hash of the block
func (b *Block) CalculateHash() string {
    // Create a struct for hashing that excludes the current hash
    hashData := struct {
        Version    int      `json:"version"`
        Index      int      `json:"index"`
        Timestamp  int64    `json:"timestamp"`
        PrevHash   string   `json:"prev_hash"`
        MerkleRoot string   `json:"merkle_root"`
        Nonce      int64    `json:"nonce"`
        Difficulty int      `json:"difficulty"`
    }{
        Version:    b.Version,
        Index:      b.Index,
        Timestamp:  b.Timestamp,
        PrevHash:   b.PrevHash,
        MerkleRoot: b.MerkleRoot,
        Nonce:      b.Nonce,
        Difficulty: b.Difficulty,
    }
    
    data, _ := json.Marshal(hashData)
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:])
}

// CalculateMerkleRoot computes the Merkle root of all transactions
func (b *Block) CalculateMerkleRoot() string {
    if len(b.Transactions) == 0 {
        return ""
    }
    
    // Simple implementation - in production, use proper Merkle tree
    var txHashes string
    for _, tx := range b.Transactions {
        txHashes += tx.Hash
    }
    
    hash := sha256.Sum256([]byte(txHashes))
    return hex.EncodeToString(hash[:])
}

// IsValid checks if the block's hash meets the difficulty requirement
func (b *Block) IsValid() bool {
    // Verify hash meets difficulty target
    hash := b.CalculateHash()
    prefix := ""
    for i := 0; i < b.Difficulty; i++ {
        prefix += "0"
    }
    
    return len(hash) >= b.Difficulty && hash[:b.Difficulty] == prefix
}

// Serialize converts the block to JSON bytes
func (b *Block) Serialize() ([]byte, error) {
    return json.Marshal(b)
}

// DeserializeBlock creates a Block from JSON bytes
func DeserializeBlock(data []byte) (*Block, error) {
    var block Block
    err := json.Unmarshal(data, &block)
    if err != nil {
        return nil, err
    }
    return &block, nil
}
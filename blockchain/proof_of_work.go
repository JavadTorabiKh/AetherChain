package blockchain

import "fmt"

// ProofOfWork implements the mining algorithm for AetherChain
type ProofOfWork struct {
    Block     *Block
    Difficulty int
}

// NewProofOfWork creates a new ProofOfWork instance
func NewProofOfWork(block *Block, difficulty int) *ProofOfWork {
    return &ProofOfWork{
        Block:     block,
        Difficulty: difficulty,
    }
}

// Mine attempts to find a valid nonce for the block
func (pow *ProofOfWork) Mine() (int64, string, error) {
    var nonce int64 = 0
    var hash string
    
    fmt.Printf("Mining block %d with difficulty %d...\n", pow.Block.Index, pow.Difficulty)
    
    for nonce < MaxNonce {
        pow.Block.Nonce = nonce
        hash = pow.Block.CalculateHash()
        
        if pow.IsValidHash(hash) {
            fmt.Printf("Block mined! Nonce: %d, Hash: %s\n", nonce, hash)
            return nonce, hash, nil
        }
        
        nonce++
    }
    
    return 0, "", fmt.Errorf("failed to mine block after %d attempts", MaxNonce)
}

// IsValidHash checks if a hash meets the difficulty requirement
func (pow *ProofOfWork) IsValidHash(hash string) bool {
    prefix := ""
    for i := 0; i < pow.Difficulty; i++ {
        prefix += "0"
    }
    
    return len(hash) >= pow.Difficulty && hash[:pow.Difficulty] == prefix
}

// Validate checks if a block's hash is valid
func (pow *ProofOfWork) Validate() bool {
    hash := pow.Block.CalculateHash()
    return pow.IsValidHash(hash)
}

// MaxNonce defines the maximum mining attempts before giving up
const MaxNonce = 100000000
package consensus

import (

	"time"
)
import "aetherchain/blockchain"

// Validator provides block and transaction validation
type Validator struct {
	blockchain *blockchain.Blockchain
}

// NewValidator creates a new validator instance
func NewValidator(bc *blockchain.Blockchain) *Validator {
	return &Validator{
		blockchain: bc,
	}
}

// ValidateTransaction validates a transaction against consensus rules
func (v *Validator) ValidateTransaction(tx *blockchain.Transaction) bool {
	// Basic transaction validation
	if !tx.IsValid() {
		return false
	}

	// Check if sender has sufficient balance
	senderBalance := v.blockchain.GetBalance(tx.From)
	if senderBalance < tx.Amount+tx.Fee {
		return false
	}

	// Check for duplicate transaction in pool
	if v.isDuplicateTransaction(tx) {
		return false
	}

	// Additional consensus rules can be added here
	// - Transaction size limits
	// - Fee requirements
	// - Script validation
	// - etc.

	return true
}

// ValidateBlock validates a block against consensus rules
func (v *Validator) ValidateBlock(block *blockchain.Block) bool {
	// Basic block structure validation
	if block == nil {
		return false
	}

	// Validate block header
	if !v.validateBlockHeader(block) {
		return false
	}

	// Validate all transactions in the block
	for _, tx := range block.Transactions {
		if !v.ValidateTransaction(tx) {
			return false
		}
	}

	// Validate block size limits (simplified)
	if len(block.Transactions) > 1000 { // Example limit
		return false
	}

	return true
}

// validateBlockHeader validates the block header
func (v *Validator) validateBlockHeader(block *blockchain.Block) bool {
	// Check block version
	if block.Version < 1 {
		return false
	}

	// Check timestamp (not too far in future)
	if block.Timestamp > time.Now().Add(2*time.Hour).Unix() {
		return false
	}

	// Check proof of work
	pow := blockchain.NewProofOfWork(block, v.blockchain.Difficulty)
	if !pow.Validate() {
		return false
	}

	// Check merkle root matches transactions
	if block.MerkleRoot != block.CalculateMerkleRoot() {
		return false
	}

	return true
}

// isDuplicateTransaction checks if a transaction already exists in the pool
func (v *Validator) isDuplicateTransaction(tx *blockchain.Transaction) bool {
	for _, existingTx := range v.blockchain.TransactionPool {
		if existingTx.Hash == tx.Hash {
			return true
		}
	}
	return false
}

// ValidateChain validates the entire blockchain
func (v *Validator) ValidateChain() bool {
	chain := v.blockchain.Chain
	
	// Check genesis block
	if len(chain) == 0 {
		return false
	}

	// Validate each block and its connection to the previous block
	for i := 1; i < len(chain); i++ {
		currentBlock := chain[i]
		previousBlock := chain[i-1]

		// Check block linkage
		if currentBlock.PrevHash != previousBlock.Hash {
			return false
		}

		// Validate current block
		if !v.ValidateBlock(currentBlock) {
			return false
		}
	}

	return true
}

// GetValidationRules returns the current validation rules
func (v *Validator) GetValidationRules() map[string]interface{} {
	return map[string]interface{}{
		"max_block_size":      1000,
		"max_transaction_fee": 1.0,
		"min_transaction_fee": 0.001,
		"allowed_versions":    []int{1},
		"difficulty":          v.blockchain.Difficulty,
	}
}
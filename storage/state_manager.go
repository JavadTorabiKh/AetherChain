package storage

import (
	"fmt"
	"sync"
	"time"

	"Aetherchain/blockchain"
)

// StateManager manages the state of the blockchain
type StateManager struct {
	blockchain *blockchain.Blockchain
	database   *Database
	mutex      sync.RWMutex
	lastSave   time.Time
}

// NewStateManager creates a new state manager
func NewStateManager(bc *blockchain.Blockchain, db *Database) *StateManager {
	return &StateManager{
		blockchain: bc,
		database:   db,
		lastSave:   time.Now(),
	}
}

// Start begins the state management service
func (sm *StateManager) Start() error {
	fmt.Println("🔄 Starting state manager...")

	// Load existing state from database
	if err := sm.database.LoadBlockchain(); err != nil {
		return fmt.Errorf("failed to load blockchain state: %v", err)
	}

	// Start periodic saving
	go sm.periodicSave()

	return nil
}

// Stop gracefully stops the state manager
func (sm *StateManager) Stop() error {
	fmt.Println("🛑 Stopping state manager...")

	// Perform final save
	if err := sm.SaveState(); err != nil {
		return fmt.Errorf("failed to save final state: %v", err)
	}

	return nil
}

// periodicSave saves state at regular intervals
func (sm *StateManager) periodicSave() {
	ticker := time.NewTicker(5 * time.Minute) // Save every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := sm.SaveState(); err != nil {
				fmt.Printf("❌ Periodic save failed: %v\n", err)
			} else {
				fmt.Println("💾 Periodic state save completed")
			}
		}
	}
}

// SaveState saves the current blockchain state
func (sm *StateManager) SaveState() error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if err := sm.database.SaveBlockchain(); err != nil {
		return err
	}

	sm.lastSave = time.Now()
	return nil
}

// GetStateInfo returns information about the current state
func (sm *StateManager) GetStateInfo() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return map[string]interface{}{
		"block_height":      len(sm.blockchain.Chain),
		"pending_txs":       len(sm.blockchain.TransactionPool),
		"accounts_count":    len(sm.blockchain.Accounts),
		"last_save":         sm.lastSave.Format(time.RFC3339),
		"time_since_save":   time.Since(sm.lastSave).String(),
		"chain_valid":       sm.blockchain.IsChainValid(),
		"total_balance":     sm.calculateTotalBalance(),
	}
}

// calculateTotalBalance calculates the total balance across all accounts
func (sm *StateManager) calculateTotalBalance() float64 {
	total := 0.0
	for _, balance := range sm.blockchain.Accounts {
		total += balance
	}
	return total
}

// AddBlock adds a block and immediately saves state
func (sm *StateManager) AddBlock(block *blockchain.Block) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Add to blockchain
	if err := sm.blockchain.AddBlock(block); err != nil {
		return err
	}

	// Save individual block
	if err := sm.database.SaveBlock(block); err != nil {
		fmt.Printf("⚠️ Failed to save individual block: %v\n", err)
	}

	// Save full state (in background to not block)
	go func() {
		if err := sm.database.SaveBlockchain(); err != nil {
			fmt.Printf("❌ Failed to save state after adding block: %v\n", err)
		}
	}()

	return nil
}

// AddTransaction adds a transaction and optionally saves state
func (sm *StateManager) AddTransaction(tx *blockchain.Transaction, saveImmediately bool) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if err := sm.blockchain.AddTransaction(tx); err != nil {
		return err
	}

	if saveImmediately {
		// Save transaction pool
		if err := sm.database.SaveBlockchain(); err != nil {
			fmt.Printf("⚠️ Failed to save state after adding transaction: %v\n", err)
		}
	}

	return nil
}

// RollbackToHeight rolls back the blockchain to a specific height
func (sm *StateManager) RollbackToHeight(height int) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if height < 0 || height >= len(sm.blockchain.Chain) {
		return fmt.Errorf("invalid height: %d", height)
	}

	fmt.Printf("↩️ Rolling back blockchain from height %d to %d\n", 
		len(sm.blockchain.Chain)-1, height)

	// Truncate chain
	sm.blockchain.Chain = sm.blockchain.Chain[:height+1]

	// Recalculate account states (simplified)
	// In production, this would rebuild state from the remaining blocks
	sm.recalculateAccountStates()

	// Save rolled back state
	if err := sm.database.SaveBlockchain(); err != nil {
		return fmt.Errorf("failed to save rolled back state: %v", err)
	}

	return nil
}

// recalculateAccountStates recalculates account balances from the current chain
func (sm *StateManager) recalculateAccountStates() {
	// Reset accounts
	sm.blockchain.Accounts = make(map[string]float64)

	// Process all blocks to rebuild account states
	for _, block := range sm.blockchain.Chain {
		for _, tx := range block.Transactions {
			sm.blockchain.Accounts[tx.From] -= tx.Amount + tx.Fee
			sm.blockchain.Accounts[tx.To] += tx.Amount
		}
		// Add miner reward
		sm.blockchain.Accounts[block.Miner] += block.BlockReward
	}
}

// GetStateSnapshot returns a snapshot of the current state
func (sm *StateManager) GetStateSnapshot() *StateSnapshot {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	snapshot := &StateSnapshot{
		BlockHeight:    len(sm.blockchain.Chain),
		LastBlockHash:  sm.blockchain.GetLastBlock().Hash,
		PendingTxCount: len(sm.blockchain.TransactionPool),
		AccountCount:   len(sm.blockchain.Accounts),
		Timestamp:      time.Now(),
	}

	return snapshot
}

// StateSnapshot represents a snapshot of the blockchain state
type StateSnapshot struct {
	BlockHeight    int       `json:"block_height"`
	LastBlockHash  string    `json:"last_block_hash"`
	PendingTxCount int       `json:"pending_tx_count"`
	AccountCount   int       `json:"account_count"`
	Timestamp      time.Time `json:"timestamp"`
}

// VerifyStateIntegrity verifies the integrity of the current state
func (sm *StateManager) VerifyStateIntegrity() (bool, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// Verify blockchain validity
	if !sm.blockchain.IsChainValid() {
		return false, fmt.Errorf("blockchain is invalid")
	}

	// Verify account balances are non-negative
	for address, balance := range sm.blockchain.Accounts {
		if balance < 0 {
			return false, fmt.Errorf("negative balance for address %s: %f", address, balance)
		}
	}

	// Verify transaction pool integrity
	for _, tx := range sm.blockchain.TransactionPool {
		if !tx.IsValid() {
			return false, fmt.Errorf("invalid transaction in pool: %s", tx.Hash)
		}
	}

	return true, nil
}
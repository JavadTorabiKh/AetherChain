package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"aetherchain/blockchain"
)

// Database handles persistent storage for the blockchain
type Database struct {
	dataDir    string
	blockchain *blockchain.Blockchain
	mutex      sync.RWMutex
}

// NewDatabase creates a new database instance
func NewDatabase(dataDir string, bc *blockchain.Blockchain) *Database {
	return &Database{
		dataDir:    dataDir,
		blockchain: bc,
	}
}

// Initialize sets up the database directory and files
func (db *Database) Initialize() error {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(db.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// Create subdirectories
	subdirs := []string{"blocks", "chainstate", "peers"}
	for _, dir := range subdirs {
		path := filepath.Join(db.dataDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create %s directory: %v", dir, err)
		}
	}

	fmt.Printf("📁 Database initialized at: %s\n", db.dataDir)
	return nil
}

// SaveBlockchain saves the entire blockchain to disk
func (db *Database) SaveBlockchain() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Save blockchain metadata
	metadata := map[string]interface{}{
		"height":        len(db.blockchain.Chain),
		"difficulty":    db.blockchain.Difficulty,
		"block_reward":  db.blockchain.BlockReward,
		"last_block":    db.blockchain.GetLastBlock().Hash,
		"genesis_block": db.blockchain.Chain[0].Hash,
	}

	if err := db.saveJSON("metadata.json", metadata); err != nil {
		return fmt.Errorf("failed to save metadata: %v", err)
	}

	// Save each block individually
	for i, block := range db.blockchain.Chain {
		filename := fmt.Sprintf("block_%d.json", i)
		if err := db.saveBlock(filename, block); err != nil {
			return fmt.Errorf("failed to save block %d: %v", i, err)
		}
	}

	// Save transaction pool
	if err := db.saveJSON("transaction_pool.json", db.blockchain.TransactionPool); err != nil {
		return fmt.Errorf("failed to save transaction pool: %v", err)
	}

	// Save account states
	if err := db.saveJSON("accounts.json", db.blockchain.Accounts); err != nil {
		return fmt.Errorf("failed to save accounts: %v", err)
	}

	fmt.Printf("💾 Blockchain saved: %d blocks, %d pending transactions\n",
		len(db.blockchain.Chain), len(db.blockchain.TransactionPool))

	return nil
}

// LoadBlockchain loads the blockchain from disk
func (db *Database) LoadBlockchain() error {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	// Check if blockchain data exists
	metadataPath := filepath.Join(db.dataDir, "metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		fmt.Println("📭 No existing blockchain data found, starting fresh")
		return nil
	}

	// Load metadata
	var metadata map[string]interface{}
	if err := db.loadJSON("metadata.json", &metadata); err != nil {
		return fmt.Errorf("failed to load metadata: %v", err)
	}

	// Load blocks
	height := int(metadata["height"].(float64))
	for i := 0; i < height; i++ {
		filename := fmt.Sprintf("block_%d.json", i)
		block, err := db.loadBlock(filename)
		if err != nil {
			return fmt.Errorf("failed to load block %d: %v", i, err)
		}
		db.blockchain.Chain = append(db.blockchain.Chain, block)
	}

	// Load transaction pool
	if err := db.loadJSON("transaction_pool.json", &db.blockchain.TransactionPool); err != nil {
		fmt.Printf("⚠️ Could not load transaction pool: %v\n", err)
	}

	// Load accounts
	if err := db.loadJSON("accounts.json", &db.blockchain.Accounts); err != nil {
		fmt.Printf("⚠️ Could not load accounts: %v\n", err)
	}

	fmt.Printf("📖 Blockchain loaded: %d blocks, %d pending transactions\n",
		len(db.blockchain.Chain), len(db.blockchain.TransactionPool))

	return nil
}

// SavePeers saves the list of known peers
func (db *Database) SavePeers(peers []string) error {
	return db.saveJSON("peers/known_peers.json", peers)
}

// LoadPeers loads the list of known peers
func (db *Database) LoadPeers() ([]string, error) {
	var peers []string
	if err := db.loadJSON("peers/known_peers.json", &peers); err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	return peers, nil
}

// SaveNodeConfig saves node configuration
func (db *Database) SaveNodeConfig(config map[string]interface{}) error {
	return db.saveJSON("node_config.json", config)
}

// LoadNodeConfig loads node configuration
func (db *Database) LoadNodeConfig() (map[string]interface{}, error) {
	var config map[string]interface{}
	if err := db.loadJSON("node_config.json", &config); err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, err
	}
	return config, nil
}

// SaveBlock saves a single block to disk
func (db *Database) SaveBlock(block *blockchain.Block) error {
	filename := fmt.Sprintf("blocks/block_%d.json", block.Index)
	return db.saveBlock(filename, block)
}

// LoadBlock loads a single block from disk
func (db *Database) LoadBlock(height int) (*blockchain.Block, error) {
	filename := fmt.Sprintf("blocks/block_%d.json", height)
	return db.loadBlock(filename)
}

// Helper methods
func (db *Database) saveJSON(filename string, data interface{}) error {
	path := filepath.Join(db.dataDir, filename)
	
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (db *Database) loadJSON(filename string, target interface{}) error {
	path := filepath.Join(db.dataDir, filename)
	
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(target)
}

func (db *Database) saveBlock(filename string, block *blockchain.Block) error {
	path := filepath.Join(db.dataDir, filename)
	
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(block)
}

func (db *Database) loadBlock(filename string) (*blockchain.Block, error) {
	path := filepath.Join(db.dataDir, filename)
	
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var block blockchain.Block
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&block); err != nil {
		return nil, err
	}

	return &block, nil
}

// GetDatabaseInfo returns database statistics
func (db *Database) GetDatabaseInfo() map[string]interface{} {
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	// Calculate total size of blockchain data
	totalSize := int64(0)
	filepath.Walk(db.dataDir, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			totalSize += info.Size()
		}
		return nil
	})

	return map[string]interface{}{
		"data_directory": db.dataDir,
		"total_size_mb":  float64(totalSize) / (1024 * 1024),
		"block_count":    len(db.blockchain.Chain),
		"tx_pool_size":   len(db.blockchain.TransactionPool),
		"account_count":  len(db.blockchain.Accounts),
	}
}
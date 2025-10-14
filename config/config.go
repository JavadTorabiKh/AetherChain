package config

import (
    "encoding/json"
    "os"
    "time"
)

// Config holds all configuration parameters for the AetherChain node
type Config struct {
    // Node Configuration
    NodeID      string `json:"node_id"`
    Version     string `json:"version"`
    Environment string `json:"environment"` // "mainnet", "testnet", "dev"
    
    // Network Configuration
    Host          string        `json:"host"`
    Port          int           `json:"port"`
    BootstrapNodes []string     `json:"bootstrap_nodes"`
    PeerTimeout   time.Duration `json:"peer_timeout"`
    
    // Blockchain Configuration
    GenesisBlockHash string  `json:"genesis_block_hash"`
    BlockReward      float64 `json:"block_reward"`
    Difficulty       int     `json:"difficulty"`
    
    // Storage Configuration
    DataDirectory string `json:"data_directory"`
    
    // API Configuration
    APIEnabled bool   `json:"api_enabled"`
    APIHost    string `json:"api_host"`
    APIPort    int    `json:"api_port"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
    return &Config{
        NodeID:          generateNodeID(),
        Version:         "1.0.0",
        Environment:     "dev",
        Host:            "0.0.0.0",
        Port:            30303,
        BootstrapNodes:  []string{},
        PeerTimeout:     30 * time.Second,
        GenesisBlockHash: "aether_genesis_2024",
        BlockReward:     50.0,
        Difficulty:      4, // Number of leading zeros required in hash
        DataDirectory:   "./data",
        APIEnabled:      true,
        APIHost:         "127.0.0.1",
        APIPort:         8080,
    }
}

// LoadConfig loads configuration from file or uses defaults
func LoadConfig(path string) (*Config, error) {
    config := DefaultConfig()
    
    if path != "" {
        file, err := os.Open(path)
        if err != nil {
            return nil, err
        }
        defer file.Close()
        
        decoder := json.NewDecoder(file)
        err = decoder.Decode(config)
        if err != nil {
            return nil, err
        }
    }
    
    return config, nil
}

func generateNodeID() string {
    // In production, this would generate a proper node ID
    return "aether_node_" + time.Now().Format("20060102150405")
}
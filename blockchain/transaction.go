package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// Transaction represents a value transfer in AetherChain
type Transaction struct {
    Version  int     `json:"version"`   // Transaction format version
    Hash     string  `json:"hash"`      // Transaction hash (ID)
    From     string  `json:"from"`      // Sender's address
    To       string  `json:"to"`        // Recipient's address
    Amount   float64 `json:"amount"`    // Amount being transferred
    Fee      float64 `json:"fee"`       // Transaction fee
    Nonce    int64   `json:"nonce"`     // Prevents replay attacks
    Timestamp int64  `json:"timestamp"` // When transaction was created
    
    // Cryptographic signatures
    Signature string `json:"signature"` // Digital signature
    PublicKey string `json:"public_key"` // Sender's public key
    
    // Metadata
    Status    string `json:"status"`    // "pending", "confirmed", "failed"
    BlockHash string `json:"block_hash"` // Hash of containing block
}

// NewTransaction creates a new transaction
func NewTransaction(from, to string, amount, fee float64, nonce int64) *Transaction {
    tx := &Transaction{
        Version:   1,
        From:      from,
        To:        to,
        Amount:    amount,
        Fee:       fee,
        Nonce:     nonce,
        Timestamp: time.Now().Unix(),
        Status:    "pending",
    }
    
    tx.Hash = tx.CalculateHash()
    return tx
}

// CalculateHash computes the transaction hash
func (tx *Transaction) CalculateHash() string {
    hashData := struct {
        Version   int     `json:"version"`
        From      string  `json:"from"`
        To        string  `json:"to"`
        Amount    float64 `json:"amount"`
        Fee       float64 `json:"fee"`
        Nonce     int64   `json:"nonce"`
        Timestamp int64   `json:"timestamp"`
    }{
        Version:   tx.Version,
        From:      tx.From,
        To:        tx.To,
        Amount:    tx.Amount,
        Fee:       tx.Fee,
        Nonce:     tx.Nonce,
        Timestamp: tx.Timestamp,
    }
    
    data, _ := json.Marshal(hashData)
    hash := sha256.Sum256(data)
    return hex.EncodeToString(hash[:])
}

// Sign creates a digital signature for the transaction
func (tx *Transaction) Sign(privateKey string) error {
    // In production, this would use proper cryptographic signing
    // For now, we'll create a simple signature
    signatureData := tx.Hash + privateKey
    hash := sha256.Sum256([]byte(signatureData))
    tx.Signature = hex.EncodeToString(hash[:])
    return nil
}

// VerifySignature checks if the transaction signature is valid
func (tx *Transaction) VerifySignature() bool {
    if tx.Signature == "" {
        return false
    }
    
    // In production, this would verify the cryptographic signature
    // For demonstration, we'll use a simple check
    expectedSignature := tx.Hash + tx.PublicKey
    hash := sha256.Sum256([]byte(expectedSignature))
    expectedHash := hex.EncodeToString(hash[:])
    
    return tx.Signature == expectedHash
}

// IsValid performs basic validation checks on the transaction
func (tx *Transaction) IsValid() bool {
    if tx.Amount <= 0 {
        return false
    }
    
    if tx.From == tx.To {
        return false
    }
    
    if !tx.VerifySignature() {
        return false
    }
    
    return true
}

// Serialize converts the transaction to JSON bytes
func (tx *Transaction) Serialize() ([]byte, error) {
    return json.Marshal(tx)
}

// DeserializeTransaction creates a Transaction from JSON bytes
func DeserializeTransaction(data []byte) (*Transaction, error) {
    var tx Transaction
    err := json.Unmarshal(data, &tx)
    if err != nil {
        return nil, err
    }
    return &tx, nil
}
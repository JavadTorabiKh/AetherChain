package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// Signer handles cryptographic signing operations
type Signer struct {
	keyManager *KeyManager
}

// NewSigner creates a new signer instance
func NewSigner(km *KeyManager) *Signer {
	return &Signer{
		keyManager: km,
	}
}

// SignData signs the given data with the specified private key
func (s *Signer) SignData(data []byte, privateKey *rsa.PrivateKey) (string, error) {
	// Hash the data
	hashed := sha256.Sum256(data)

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %v", err)
	}

	// Return hex-encoded signature
	return hex.EncodeToString(signature), nil
}

// VerifySignature verifies a signature against the given data and public key
func (s *Signer) VerifySignature(data []byte, signature string, publicKey *rsa.PublicKey) bool {
	// Decode signature
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	// Hash the data
	hashed := sha256.Sum256(data)

	// Verify the signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed[:], sigBytes)
	return err == nil
}

// SignTransaction signs a transaction with the given private key
func (s *Signer) SignTransaction(txData []byte, privateKey *rsa.PrivateKey) (string, error) {
	return s.SignData(txData, privateKey)
}

// VerifyTransactionSignature verifies a transaction signature
func (s *Signer) VerifyTransactionSignature(txData []byte, signature string, publicKey *rsa.PublicKey) bool {
	return s.VerifySignature(txData, signature, publicKey)
}

// GenerateSignature creates a signature for a message
func (s *Signer) GenerateSignature(message string, keyName string) (string, error) {
	// Load key pair
	keyPair, err := s.keyManager.LoadKeyPair(keyName)
	if err != nil {
		return "", fmt.Errorf("failed to load key pair: %v", err)
	}

	// Sign the message
	signature, err := s.SignData([]byte(message), keyPair.PrivateKey)
	if err != nil {
		return "", err
	}

	return signature, nil
}

// VerifyMessageSignature verifies a message signature
func (s *Signer) VerifyMessageSignature(message, signature string, publicKey *rsa.PublicKey) bool {
	return s.VerifySignature([]byte(message), signature, publicKey)
}

// GetPublicKeyFromSignature recovers public key information from signature (placeholder)
// Note: In real RSA, you can't recover public key from signature alone
// This is a simplified version for demonstration
func (s *Signer) GetPublicKeyFromSignature(data []byte, signature string) (*rsa.PublicKey, error) {
	// In a real implementation, this would require additional context
	// For now, return an error as this is not straightforward with RSA
	return nil, fmt.Errorf("public key recovery not supported with RSA")
}

// SignatureInfo represents information about a signature
type SignatureInfo struct {
	Algorithm string `json:"algorithm"`
	Hash      string `json:"hash"`
	KeySize   int    `json:"key_size"`
	Address   string `json:"address"`
}

// GetSignatureInfo returns information about a signature
func (s *Signer) GetSignatureInfo(data []byte, signature string, publicKey *rsa.PublicKey) (*SignatureInfo, error) {
	// Verify signature first
	if !s.VerifySignature(data, signature, publicKey) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Calculate data hash
	hashed := sha256.Sum256(data)

	// Generate address from public key
	address := s.keyManager.GetAddressFromPublicKey(publicKey)

	return &SignatureInfo{
		Algorithm: "RSA-SHA256",
		Hash:      hex.EncodeToString(hashed[:]),
		KeySize:   2048,
		Address:   address,
	}, nil
}

// BatchVerify verifies multiple signatures in batch (placeholder)
func (s *Signer) BatchVerify(verifications []VerificationRequest) ([]bool, error) {
	results := make([]bool, len(verifications))
	
	for i, req := range verifications {
		results[i] = s.VerifySignature(req.Data, req.Signature, req.PublicKey)
	}
	
	return results, nil
}

// VerificationRequest represents a signature verification request
type VerificationRequest struct {
	Data      []byte
	Signature string
	PublicKey *rsa.PublicKey
}

// CreateDetachedSignature creates a detached signature package
func (s *Signer) CreateDetachedSignature(data []byte, privateKey *rsa.PrivateKey) (map[string]interface{}, error) {
	signature, err := s.SignData(data, privateKey)
	if err != nil {
		return nil, err
	}

	// Calculate data hash
	hashed := sha256.Sum256(data)

	return map[string]interface{}{
		"signature": signature,
		"data_hash": hex.EncodeToString(hashed[:]),
		"algorithm": "RSA-SHA256",
		"timestamp": time.Now().Unix(),
	}, nil
}

// VerifyDetachedSignature verifies a detached signature
func (s *Signer) VerifyDetachedSignature(data []byte, signaturePackage map[string]interface{}, publicKey *rsa.PublicKey) bool {
	signature, ok := signaturePackage["signature"].(string)
	if !ok {
		return false
	}

	return s.VerifySignature(data, signature, publicKey)
}
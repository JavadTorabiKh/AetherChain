package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
)

// KeyManager handles cryptographic key generation and management
type KeyManager struct {
	keysDir string
}

// NewKeyManager creates a new key manager
func NewKeyManager(keysDir string) *KeyManager {
	return &KeyManager{
		keysDir: keysDir,
	}
}

// KeyPair represents a public/private key pair
type KeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Address    string
}

// GenerateKeyPair generates a new RSA key pair
func (km *KeyManager) GenerateKeyPair() (*KeyPair, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %v", err)
	}

	// Create key pair structure
	keyPair := &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}

	// Generate address from public key
	keyPair.Address = km.generateAddress(keyPair.PublicKey)

	return keyPair, nil
}

// generateAddress creates a blockchain address from a public key
func (km *KeyManager) generateAddress(publicKey *rsa.PublicKey) string {
	// Serialize public key
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return ""
	}

	// Hash the public key
	hash := sha256.Sum256(publicKeyBytes)

	// Take first 20 bytes for address (similar to Ethereum)
	addressBytes := hash[:20]

	// Convert to hex string
	return "0x" + hex.EncodeToString(addressBytes)
}

// SaveKeyPair saves a key pair to disk
func (km *KeyManager) SaveKeyPair(keyPair *KeyPair, name string) error {
	// Create keys directory if it doesn't exist
	if err := os.MkdirAll(km.keysDir, 0700); err != nil {
		return fmt.Errorf("failed to create keys directory: %v", err)
	}

	// Save private key
	privateKeyPath := filepath.Join(km.keysDir, name+".key")
	if err := km.savePrivateKey(keyPair.PrivateKey, privateKeyPath); err != nil {
		return fmt.Errorf("failed to save private key: %v", err)
	}

	// Save public key
	publicKeyPath := filepath.Join(km.keysDir, name+".pub")
	if err := km.savePublicKey(keyPair.PublicKey, publicKeyPath); err != nil {
		return fmt.Errorf("failed to save public key: %v", err)
	}

	// Save address
	addressPath := filepath.Join(km.keysDir, name+".address")
	if err := os.WriteFile(addressPath, []byte(keyPair.Address), 0600); err != nil {
		return fmt.Errorf("failed to save address: %v", err)
	}

	fmt.Printf("🔑 Key pair saved: %s (Address: %s)\n", name, keyPair.Address)
	return nil
}

// LoadKeyPair loads a key pair from disk
func (km *KeyManager) LoadKeyPair(name string) (*KeyPair, error) {
	// Load private key
	privateKeyPath := filepath.Join(km.keysDir, name+".key")
	privateKey, err := km.loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %v", err)
	}

	// Create key pair
	keyPair := &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
	}

	// Load or generate address
	addressPath := filepath.Join(km.keysDir, name+".address")
	if addressData, err := os.ReadFile(addressPath); err == nil {
		keyPair.Address = string(addressData)
	} else {
		// Generate address if not saved
		keyPair.Address = km.generateAddress(keyPair.PublicKey)
	}

	return keyPair, nil
}

// savePrivateKey saves an RSA private key to a file
func (km *KeyManager) savePrivateKey(privateKey *rsa.PrivateKey, path string) error {
	// Encode private key to PEM format
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return os.WriteFile(path, privateKeyPEM, 0600)
}

// savePublicKey saves an RSA public key to a file
func (km *KeyManager) savePublicKey(publicKey *rsa.PublicKey, path string) error {
	// Encode public key to PEM format
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return err
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return os.WriteFile(path, publicKeyPEM, 0644)
}

// loadPrivateKey loads an RSA private key from a file
func (km *KeyManager) loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decode PEM data
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("invalid private key format")
	}

	// Parse private key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// GetAddressFromPublicKey generates an address from a public key
func (km *KeyManager) GetAddressFromPublicKey(publicKey *rsa.PublicKey) string {
	return km.generateAddress(publicKey)
}

// ListKeys returns a list of all saved key pairs
func (km *KeyManager) ListKeys() ([]string, error) {
	files, err := os.ReadDir(km.keysDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var keys []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".key" {
			keys = append(keys, file.Name()[:len(file.Name())-4])
		}
	}

	return keys, nil
}

// KeyExists checks if a key pair with the given name exists
func (km *KeyManager) KeyExists(name string) bool {
	privateKeyPath := filepath.Join(km.keysDir, name+".key")
	_, err := os.Stat(privateKeyPath)
	return err == nil
}

// GetKeyInfo returns information about a key pair
func (km *KeyManager) GetKeyInfo(name string) (map[string]interface{}, error) {
	keyPair, err := km.LoadKeyPair(name)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":       name,
		"address":    keyPair.Address,
		"key_size":   2048,
		"algorithm":  "RSA",
		"public_key": hex.EncodeToString(x509.MarshalPKCS1PublicKey(keyPair.PublicKey)),
	}, nil
}
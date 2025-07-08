package mcp

import (
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Wallet represents an Ethereum wallet.
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

// NewWallet creates a new Ethereum wallet.
func NewWallet() (*Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &Wallet{
		PrivateKey: privateKey,
		Address:    address,
	}, nil
}

// SaveWallet saves the wallet to a file.
func SaveWallet(wallet *Wallet, path string, password string) error {
	key := &keystore.Key{
		Address:    wallet.Address,
		PrivateKey: wallet.PrivateKey,
	}
	keyjson, err := keystore.EncryptKey(key, password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return fmt.Errorf("failed to encrypt key: %w", err)
	}
	return os.WriteFile(path, keyjson, 0600)
}

// LoadWallet loads a wallet from a file.
func LoadWallet(path string, password string) (*Wallet, error) {
	keyjson, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key: %w", err)
	}
	return &Wallet{
		PrivateKey: key.PrivateKey,
		Address:    key.Address,
	}, nil
}

// GetOrCreateWallet loads a wallet from a file or creates a new one if it doesn't exist.
func GetOrCreateWallet(path string, password string) (*Wallet, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		wallet, err := NewWallet()
		if err != nil {
			return nil, err
		}
		if err := SaveWallet(wallet, path, password); err != nil {
			return nil, err
		}
		return wallet, nil
	}
	return LoadWallet(path, password)
}

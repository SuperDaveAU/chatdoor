package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
)

const (
	PublicKeySize  = 32
	PrivateKeySize = 32
	NonceSize      = 24
)

// KeyPair holds a Curve25519 keypair
type KeyPair struct {
	publicKey  [PublicKeySize]byte
	privateKey [PrivateKeySize]byte
}

// GenerateKeyPair creates a new Curve25519 keypair
func GenerateKeyPair() (*KeyPair, error) {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating keypair: %w", err)
	}

	return &KeyPair{
		publicKey:  *publicKey,
		privateKey: *privateKey,
	}, nil
}

// PublicKey returns the public key
func (kp *KeyPair) PublicKey() []byte {
	return kp.publicKey[:]
}

// PrivateKey returns the private key
func (kp *KeyPair) PrivateKey() []byte {
	return kp.privateKey[:]
}

// Encrypt encrypts a message for a recipient's public key
func (kp *KeyPair) Encrypt(message []byte, recipientPublicKey []byte) ([]byte, error) {
	if len(recipientPublicKey) != PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
	}

	var nonce [NonceSize]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	var recipientPubKey [PublicKeySize]byte
	copy(recipientPubKey[:], recipientPublicKey)

	encrypted := box.Seal(nonce[:], message, &nonce, &recipientPubKey, &kp.privateKey)
	return encrypted, nil
}

// Decrypt decrypts a message from a sender's public key
func (kp *KeyPair) Decrypt(encrypted []byte, senderPublicKey []byte) ([]byte, error) {
	if len(encrypted) < NonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	if len(senderPublicKey) != PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
	}

	var nonce [NonceSize]byte
	copy(nonce[:], encrypted[:NonceSize])

	var senderPubKey [PublicKeySize]byte
	copy(senderPubKey[:], senderPublicKey)

	decrypted, ok := box.Open(nil, encrypted[NonceSize:], &nonce, &senderPubKey, &kp.privateKey)
	if !ok {
		return nil, fmt.Errorf("decryption failed")
	}

	return decrypted, nil
}

// SavePrivateKey saves the private key to a file
func SavePrivateKey(kp *KeyPair, path string) error {
	return os.WriteFile(path, kp.privateKey[:], 0600)
}

// LoadPrivateKey loads a private key from a file
func LoadPrivateKey(path string) (*KeyPair, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	if len(data) != PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size")
	}

	var privateKey [PrivateKeySize]byte
	copy(privateKey[:], data)

	// Derive public key from private key
	var publicKey [PublicKeySize]byte
	curve25519.ScalarBaseMult(&publicKey, &privateKey)

	return &KeyPair{
		publicKey:  publicKey,
		privateKey: privateKey,
	}, nil
}

// SavePublicKey saves the public key to a file
func SavePublicKey(publicKey []byte, path string) error {
	return os.WriteFile(path, publicKey, 0644)
}

// LoadPublicKeyFromFile loads a public key from a file
func LoadPublicKeyFromFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading public key: %w", err)
	}

	if len(data) != PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
	}

	return data, nil
}

// PublicKeyToBase64 converts a public key to base64 string
func PublicKeyToBase64(publicKey []byte) string {
	return base64.StdEncoding.EncodeToString(publicKey)
}

// Base64ToPublicKey converts a base64 string to public key
func Base64ToPublicKey(b64 string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	if len(data) != PublicKeySize {
		return nil, fmt.Errorf("invalid public key size")
	}

	return data, nil
}

// ValidatePublicKey checks if a public key is valid
func ValidatePublicKey(publicKey []byte) bool {
	return len(publicKey) == PublicKeySize
}

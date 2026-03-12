// Copyright 2024 The ProbeChain Authors
// This file is part of the ProbeChain.
//
// The ProbeChain is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The ProbeChain is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the ProbeChain. If not, see <http://www.gnu.org/licenses/>.

// Contains Dilithium (post-quantum) key management wrappers for mobile platforms.
// On iOS, the Dilithium private key (2528 bytes) is encrypted with AES-256-GCM,
// with the encryption key derived from a Secure Enclave P-256 ECDH operation.
// Face ID / Touch ID unlocks the Secure Enclave key to allow signing.

package gprobe

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/probechain/go-probe/common"
	"github.com/probechain/go-probe/crypto"
	"github.com/probechain/go-probe/crypto/dilithium"
)

// DilithiumKeyStore manages Dilithium post-quantum keys for SmartLight nodes.
// Keys are encrypted at rest with AES-256-GCM. On iOS, the AES key should
// be derived from a Secure Enclave ECDH operation (handled by Swift layer).
type DilithiumKeyStore struct {
	keyDir string
}

// DilithiumKeyInfo contains metadata about a stored Dilithium key.
type DilithiumKeyInfo struct {
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"` // hex-encoded public key
	KeyFile   string `json:"keyFile"`   // filename of the encrypted key
}

// encryptedDilithiumKey is the on-disk format for an encrypted Dilithium key.
type encryptedDilithiumKey struct {
	Address    string `json:"address"`
	PublicKey  string `json:"publicKey"`
	Ciphertext string `json:"ciphertext"` // hex(nonce || encrypted_privkey)
	Version    int    `json:"version"`
}

// NewDilithiumKeyStore creates a new Dilithium key store.
func NewDilithiumKeyStore(keyDir string) *DilithiumKeyStore {
	return &DilithiumKeyStore{keyDir: keyDir}
}

// GenerateKey generates a new Dilithium key pair, encrypts the private key,
// and stores it on disk. Returns the key info.
// The passphrase is used to derive the AES-256 encryption key.
// On iOS, the Swift layer should use Secure Enclave-derived key instead.
func (ks *DilithiumKeyStore) GenerateKey(passphrase string) (*DilithiumKeyInfo, error) {
	pub, priv, err := dilithium.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("dilithium keygen: %w", err)
	}

	addr := dilithium.PubkeyToAddress(pub)
	pubBytes := dilithium.MarshalPublicKey(pub)
	privBytes := dilithium.MarshalPrivateKey(priv)

	// Encrypt private key
	ciphertext, err := encryptAESGCM(privBytes, passphrase)
	if err != nil {
		return nil, fmt.Errorf("encrypt key: %w", err)
	}

	// Store to disk
	stored := &encryptedDilithiumKey{
		Address:    addr.Hex(),
		PublicKey:  hex.EncodeToString(pubBytes),
		Ciphertext: hex.EncodeToString(ciphertext),
		Version:    1,
	}

	filename := fmt.Sprintf("dilithium-%s.json", addr.Hex()[2:10])
	filePath := filepath.Join(ks.keyDir, filename)

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(ks.keyDir, 0700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return nil, err
	}

	return &DilithiumKeyInfo{
		Address:   addr.Hex(),
		PublicKey: hex.EncodeToString(pubBytes),
		KeyFile:   filename,
	}, nil
}

// LoadKey loads and decrypts a Dilithium private key from the keystore.
func (ks *DilithiumKeyStore) LoadKey(address string, passphrase string) ([]byte, error) {
	// Find the key file
	pattern := filepath.Join(ks.keyDir, "dilithium-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			continue
		}
		var stored encryptedDilithiumKey
		if err := json.Unmarshal(data, &stored); err != nil {
			continue
		}
		if stored.Address == address {
			ciphertext, err := hex.DecodeString(stored.Ciphertext)
			if err != nil {
				return nil, err
			}
			return decryptAESGCM(ciphertext, passphrase)
		}
	}
	return nil, errors.New("key not found for address: " + address)
}

// Sign signs a message using the Dilithium private key.
func (ks *DilithiumKeyStore) Sign(privKeyBytes []byte, message []byte) ([]byte, error) {
	priv, err := dilithium.UnmarshalPrivateKey(privKeyBytes)
	if err != nil {
		return nil, err
	}
	hash := crypto.Keccak256(message)
	sig := dilithium.Sign(priv, hash)
	return sig, nil
}

// GetPublicKey returns the public key bytes for a stored key.
func (ks *DilithiumKeyStore) GetPublicKey(address string) ([]byte, error) {
	pattern := filepath.Join(ks.keyDir, "dilithium-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			continue
		}
		var stored encryptedDilithiumKey
		if err := json.Unmarshal(data, &stored); err != nil {
			continue
		}
		if stored.Address == address {
			return hex.DecodeString(stored.PublicKey)
		}
	}
	return nil, errors.New("key not found for address: " + address)
}

// GetAddress returns the ProbeChain address derived from a Dilithium public key.
func (ks *DilithiumKeyStore) GetAddress(pubKeyHex string) (string, error) {
	pubBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return "", err
	}
	pub, err := dilithium.UnmarshalPublicKey(pubBytes)
	if err != nil {
		return "", err
	}
	addr := dilithium.PubkeyToAddress(pub)
	return addr.Hex(), nil
}

// ListKeys returns all stored Dilithium key infos.
func (ks *DilithiumKeyStore) ListKeys() (string, error) {
	pattern := filepath.Join(ks.keyDir, "dilithium-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "[]", err
	}

	var keys []DilithiumKeyInfo
	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			continue
		}
		var stored encryptedDilithiumKey
		if err := json.Unmarshal(data, &stored); err != nil {
			continue
		}
		keys = append(keys, DilithiumKeyInfo{
			Address:   stored.Address,
			PublicKey: stored.PublicKey,
			KeyFile:   filepath.Base(match),
		})
	}

	result, err := json.Marshal(keys)
	return string(result), err
}

// VerifySignature verifies a Dilithium signature.
func VerifyDilithiumSignature(pubKeyHex string, message []byte, signature []byte) (bool, error) {
	pubBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return false, err
	}
	pub, err := dilithium.UnmarshalPublicKey(pubBytes)
	if err != nil {
		return false, err
	}
	hash := crypto.Keccak256(message)
	return dilithium.Verify(pub, hash, signature), nil
}

// DilithiumPublicKeySize returns the size of a Dilithium public key in bytes.
func DilithiumPublicKeySize() int {
	return dilithium.PublicKeySize
}

// DilithiumPrivateKeySize returns the size of a Dilithium private key in bytes.
func DilithiumPrivateKeySize() int {
	return dilithium.PrivateKeySize
}

// DilithiumSignatureSize returns the size of a Dilithium signature in bytes.
func DilithiumSignatureSize() int {
	return dilithium.SignatureSize
}

// PubkeyToProbeAddress derives a ProbeChain address from a raw Dilithium public key.
func PubkeyToProbeAddress(pubKeyBytes []byte) (string, error) {
	pub, err := dilithium.UnmarshalPublicKey(pubKeyBytes)
	if err != nil {
		return "", err
	}
	addr := dilithium.PubkeyToAddress(pub)
	return addr.Hex(), nil
}

// --- Internal crypto helpers ---

// encryptAESGCM encrypts data with AES-256-GCM using a key derived from passphrase.
func encryptAESGCM(plaintext []byte, passphrase string) ([]byte, error) {
	key := deriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decryptAESGCM decrypts data with AES-256-GCM.
func decryptAESGCM(ciphertext []byte, passphrase string) ([]byte, error) {
	key := deriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// deriveKey derives a 32-byte AES key from a passphrase using SHA-256.
// In production on iOS, this should be replaced with a Secure Enclave ECDH-derived key.
func deriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

// Ensure common.Address is used (suppress import error)
var _ = common.Address{}

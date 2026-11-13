package model

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/Mopsgamer/draqun/server/environment"
)

const encPrefix = "enc:v1:"

func getEncryptionKey() []byte {
	key := environment.JWTKey
	if key == "" {
		key = "default-draqun-insecure-secret-key-change-in-prod"
	}
	hash := sha256.Sum256([]byte(key + ":db-encryption-salt"))
	return hash[:]
}

// EncryptStringDeterministically encrypts a plaintext string deterministically using AES-256-GCM
// where the nonce is derived via HMAC-SHA256 from the plaintext.
// This allows identical plaintexts to yield identical ciphertexts, enabling exact match lookups
// (e.g. WHERE email = ?) and UNIQUE database constraints on encrypted values.
func EncryptStringDeterministically(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	if strings.HasPrefix(plain, encPrefix) {
		return plain, nil
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Deterministic nonce derived from HMAC-SHA256(key, plaintext) truncated to GCM nonce size
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(plain))
	nonceHash := mac.Sum(nil)
	nonce := nonceHash[:gcm.NonceSize()]

	ciphertext := gcm.Seal(nil, nonce, []byte(plain), nil)
	combined := append(nonce, ciphertext...)

	return encPrefix + base64.RawURLEncoding.EncodeToString(combined), nil
}

// DecryptStringDeterministically decrypts a ciphertext produced by EncryptStringDeterministically.
// If the string does not have the encrypted prefix, it returns it as is (supporting unencrypted legacy data).
func DecryptStringDeterministically(encoded string) (string, error) {
	if !strings.HasPrefix(encoded, encPrefix) {
		return encoded, nil
	}

	raw := strings.TrimPrefix(encoded, encPrefix)
	combined, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return encoded, err
	}

	key := getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return encoded, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return encoded, err
	}

	nonceSize := gcm.NonceSize()
	if len(combined) < nonceSize {
		return encoded, nil
	}

	nonce := combined[:nonceSize]
	ciphertext := combined[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return encoded, err
	}

	return string(plaintext), nil
}

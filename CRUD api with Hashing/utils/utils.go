package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// ------------------
// Argon2 helper (encode/decode)
// ------------------

// Parameters -- tune for your environment
var (
	ArgonTime    uint32 = 1
	ArgonMemory  uint32 = 64 * 1024 // 64 MB
	ArgonThreads uint8  = 4
	ArgonKeyLen  uint32 = 32
	SaltLen             = 16
)

func generateSalt(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// HashPassword returns an encoded hash that contains the parameters, salt and derived key
func HashPassword(password string) (string, error) {
	salt, err := generateSalt(SaltLen)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemory, ArgonThreads, ArgonKeyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", ArgonMemory, ArgonTime, ArgonThreads, b64Salt, b64Hash)
	return encoded, nil
}

// ComparePassword verifies password against encoded hash
func ComparePassword(encodedHash, password string) (bool, error) {
	// Expected format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	paramsPart := parts[3]
	saltB64 := parts[4]
	hashB64 := parts[5]

	var memory uint32
	var timeCost uint32
	var threads uint8
	_, err := fmt.Sscanf(paramsPart, "m=%d,t=%d,p=%d", &memory, &timeCost, &threads)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, err
	}

	calcHash := argon2.IDKey([]byte(password), salt, timeCost, memory, threads, uint32(len(hash)))

	if subtleConstantTimeCompare(calcHash, hash) {
		return true, nil
	}
	return false, nil
}

// Constant-time comparison
func subtleConstantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte = 0
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

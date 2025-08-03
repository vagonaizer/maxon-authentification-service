package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateRandomBytes(length int) ([]byte, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	return bytes, nil
}

func HashSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func GenerateSecureToken() (string, error) {
	return GenerateRandomString(32)
}

func GenerateAPIKey() (string, error) {
	prefix := "ak_"
	randomPart, err := GenerateRandomString(32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", prefix, randomPart), nil
}

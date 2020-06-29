package secrets

import (
	"crypto/rand"
	"fmt"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789$:#"

// GenerateString creates a random secret of the desired length
func GenerateString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("Failed to generate secret of desired length %d", length)
	}
	s := make([]byte, length)
	for i, v := range b {
		s[i] = charset[int(v)%len(charset)]
	}
	return string(s), nil
}

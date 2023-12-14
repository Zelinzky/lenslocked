package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func Bytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	read, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("bytes: %w", err)
	}
	if read < n {
		return nil, fmt.Errorf("bytes: didn't read enough random bytes")
	}
	return bytes, nil
}

func String(n int) (string, error) {
	bytes, err := Bytes(n)
	if err != nil {
		return "", fmt.Errorf("string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

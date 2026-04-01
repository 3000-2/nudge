package idgen

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	charset     = "abcdefghijklmnopqrstuvwxyz0123456789"
	idLength    = 8
	maxAttempts = 256
)

func GenerateUnique(exists func(string) (bool, error)) (string, error) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		id, err := generate()
		if err != nil {
			return "", err
		}

		collision, err := exists(id)
		if err != nil {
			return "", err
		}
		if !collision {
			return id, nil
		}
	}

	return "", fmt.Errorf("generate unique reminder id: exhausted %d attempts", maxAttempts)
}

func generate() (string, error) {
	buf := make([]byte, idLength)
	limit := big.NewInt(int64(len(charset)))

	for i := range buf {
		n, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("generate reminder id: %w", err)
		}
		buf[i] = charset[n.Int64()]
	}

	return string(buf), nil
}

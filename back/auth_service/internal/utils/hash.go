package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenerateHash(str string) (string, error) {
	hash := sha256.New()

	if _, err := hash.Write([]byte(str)); err != nil {
		return "", err
	}

	hashBytes := hash.Sum(nil)
	return hex.EncodeToString(hashBytes), nil
}

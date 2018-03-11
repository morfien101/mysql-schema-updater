package md5er

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// Hash Converts a string into a md5 hash
func Hash(text string) (string, error) {
	hasher := md5.New()
	_, err := hasher.Write([]byte(text))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// HashFile will read a file and hash it.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

func Rand(len int) ([]byte, error) {
	salt := make([]byte, len)
	if n, err := rand.Read(salt); err != nil || n != len {
		return nil, fmt.Errorf("failed to generate salt: %v", err)
	}
	return salt, nil
}

func GenChecksumSHA(path string) (chksm string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	sum := hash.Sum(nil)

	return base64.StdEncoding.EncodeToString(sum), nil
}

func GenChecksumSHABytes(data []byte) (chksm string, err error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, bytes.NewReader(data)); err != nil {
		return "", err
	}
	sum := hash.Sum(nil)

	return base64.StdEncoding.EncodeToString(sum), nil
}

func GenChecksumHMAC(path string, key []byte) (chksm string, err error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, key)
	if _, err := io.Copy(mac, file); err != nil {
		return "", err
	}
	sum := mac.Sum(nil)

	return base64.StdEncoding.EncodeToString(sum), nil

}

func GenChecksumHMACBytes(data, key []byte) (chksm string, err error) {
	mac := hmac.New(sha256.New, key)
	if _, err = mac.Write(data); err != nil {
		return "", err
	}
	sum := mac.Sum(nil)

	return base64.StdEncoding.EncodeToString(sum), nil
}

package rand

import (
	"crypto/rand"
	"encoding/base64"
)

// RememberTokenBytes is the length of the byte slice used for generating the remember token.
const RememberTokenBytes = 32

// Bytes will generate n random bytes or return an error.
// This uses the crypto/rand package so is safe for
// generating tokens.
func Bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// NBytes returns the numbver of bytes used in teh base64 URL encoded string
func NBytes(base64String string) (int, error) {
	b, err := base64.URLEncoding.DecodeString(base64String)
	if err != nil {
		return -1, err
	}
	return len(b), nil
}

// String returns a URLencoded string representation of a byte slice of nBytes length.
func String(nBytes int) (string, error) {
	b, err := Bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// RememberToken is a helper function to generate remember tokens
// of a predetermined size.
func RememberToken() (string, error) {
	return String(RememberTokenBytes)
}

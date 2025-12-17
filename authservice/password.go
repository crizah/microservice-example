package authservice

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, string, error) {

	const (
		salt_bytes = 16
		hash_bytes = 32
		time       = 3
		memory     = 64 * 1024
		threads    = 4
	)

	// generate salt
	salt := make([]byte, salt_bytes)

	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", "", errors.New("Err generating random salt")
	}

	// generate hash with argon2

	key := argon2.IDKey([]byte(password), salt, time, memory, threads, hash_bytes)

	// return salt and hash
	return base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
		nil

}

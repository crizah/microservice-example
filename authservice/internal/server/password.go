package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	SALT_BYTES = 16
	HASH_BYTES = 32
	TIME       = 3
	MEMORY     = 64 * 1024
	THREADS    = 4
)

func HashPassword(password string) (string, string, error) {

	// generate salt
	salt := make([]byte, SALT_BYTES)

	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", "", errors.New("err generating random salt")
	}

	// generate hash with argon2

	key := argon2.IDKey([]byte(password), salt, TIME, MEMORY, THREADS, HASH_BYTES)

	// return salt and hash
	return base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
		nil

}

func VerifyPass(password string, salt string, hash string) (bool, error) {
	// recreate hash from password and salt
	// compare, if same, verified
	saltbytes, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return false, err
	}

	hashBytes, err := base64.RawStdEncoding.DecodeString(hash)
	if err != nil {
		return false, err
	}

	key := argon2.IDKey([]byte(password), saltbytes, TIME, MEMORY, THREADS, HASH_BYTES)

	if subtle.ConstantTimeCompare(key, hashBytes) == 1 {
		return true, nil

	}
	return false, nil

}

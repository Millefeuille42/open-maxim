package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"

	"golang.org/x/crypto/argon2"
)

// https://snyk.io/blog/secure-password-hashing-in-go/

var ErrPasswordMismatch = errors.New("wrong username or password")

type Argon2idHash struct {
	// time represents the number of
	// passed over the specified memory.
	time uint32
	// cpu memory to be used.
	memory uint32
	// threads for parallelism aspect
	// of the algorithm.
	threads uint8
	// keyLen of the generate hash key.
	keyLen uint32
	// saltLen the length of the salt used.
	saltLen uint32
}

// HashSalt struct used to store
// generated hash and salt used to
// generate the hash.
type HashSalt struct {
	Hash, Salt []byte
	Params     Argon2idHash
}

// NewArgon2idHash constructor function for
// Argon2idHash.
func NewArgon2idHash(time, saltLen uint32, memory uint32, threads uint8, keyLen uint32) *Argon2idHash {
	return &Argon2idHash{
		time:    time,
		saltLen: saltLen,
		memory:  memory,
		threads: threads,
		keyLen:  keyLen,
	}
}

func randomSecret(length uint32) ([]byte, error) {
	secret := make([]byte, length)

	_, err := rand.Read(secret)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

// GenerateHash using the password and the generated salt.
func (a *Argon2idHash) GenerateHash(password []byte) (*HashSalt, error) {
	// Generate a salt of the configured salt length.
	salt, err := randomSecret(a.saltLen)
	if err != nil {
		return nil, err
	}
	// Generate hash
	hash := argon2.IDKey(password, salt, a.time, a.memory, a.threads, a.keyLen)
	// Return the generated hash and salt used for storage.
	return &HashSalt{Hash: hash, Salt: salt, Params: *a}, nil
}

// Compare generated hash with store hash.
func (a *Argon2idHash) Compare(hashSalt HashSalt, password []byte) error {
	p := hashSalt.Params
	// Generate hash for comparison.
	hash := argon2.IDKey(password, hashSalt.Salt, p.time, p.memory, p.threads, p.keyLen)
	// Compare the generated hash with the stored hash.
	// If they don't match return error.
	if subtle.ConstantTimeCompare(hashSalt.Hash, hash) != 1 {
		return ErrPasswordMismatch
	}
	return nil
}

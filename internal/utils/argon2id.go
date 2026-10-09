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
	Time uint32
	// cpu memory to be used.
	Memory uint32
	// threads for parallelism aspect
	// of the algorithm.
	Threads uint8
	// keyLen of the generate hash key.
	KeyLen uint32
	// saltLen the length of the salt used.
	SaltLen uint32
}

// HashSalt struct used to store
// generated hash and salt used to
// generate the hash.
type HashSalt struct {
	Hash, Salt []byte
	Params     Argon2idHash
}

func DefaultArgon2idParams() Argon2idHash {
	return Argon2idHash{Time: 1, Memory: 64 * 1024, Threads: 4, KeyLen: 32, SaltLen: 16}
}

// NewArgon2idHash constructor function for
// Argon2idHash.
func NewArgon2idHash(time, saltLen uint32, memory uint32, threads uint8, keyLen uint32) *Argon2idHash {
	return &Argon2idHash{
		Time:    time,
		SaltLen: saltLen,
		Memory:  memory,
		Threads: threads,
		KeyLen:  keyLen,
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
	salt, err := randomSecret(a.SaltLen)
	if err != nil {
		return nil, err
	}
	// Generate hash
	hash := argon2.IDKey(password, salt, a.Time, a.Memory, a.Threads, a.KeyLen)
	// Return the generated hash and salt used for storage.
	return &HashSalt{Hash: hash, Salt: salt, Params: *a}, nil
}

// Compare generated hash with store hash.
func (a *Argon2idHash) Compare(hashSalt HashSalt, password []byte) error {
	p := hashSalt.Params
	// Generate hash for comparison.
	hash := argon2.IDKey(password, hashSalt.Salt, p.Time, p.Memory, p.Threads, p.KeyLen)
	// Compare the generated hash with the stored hash.
	// If they don't match return error.
	if subtle.ConstantTimeCompare(hashSalt.Hash, hash) != 1 {
		return ErrPasswordMismatch
	}
	return nil
}

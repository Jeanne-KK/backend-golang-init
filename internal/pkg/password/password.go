package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"backend/internal/config"

	"golang.org/x/crypto/argon2"
)

const phcPrefix = "$argon2id$v=19$m="

type Hasher struct {
	cfg config.Argon2Config
}

func NewHasher(cfg config.Argon2Config) *Hasher {
	return &Hasher{cfg: cfg}
}

func (h *Hasher) Hash(password string) (string, error) {
	//	random salt
	salt := make([]byte, h.cfg.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	//	hash the password
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.cfg.Iterations,
		h.cfg.MemoryKiB,
		h.cfg.Threads,
		h.cfg.KeyLength,
	)

	//	encode the salt and hash
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	//	format the phc password hash
	encoded := fmt.Sprintf("%s%d,t=%d,p=%d$%s$%s",
		phcPrefix,
		h.cfg.MemoryKiB,
		h.cfg.Iterations,
		h.cfg.Threads,
		b64Salt,
		b64Hash,
	)
	return encoded, nil
}

func (h *Hasher) Verify(password, encoded string) (bool, error) {
	//	get the phc decoded
	params, salt, hash, err := decodePHC(encoded)
	if err != nil {
		return false, err
	}

	other := argon2.IDKey([]byte(password), salt, params.iterations, params.memoryKiB, params.threads, uint32(len(hash)))
	//	compare the hash and the other hash
	if subtle.ConstantTimeCompare(hash, other) == 1 {
		return true, nil
	}
	//	return false if the hash and the other hash are not the same
	return false, nil
}

type phcParams struct {
	memoryKiB  uint32
	iterations uint32
	threads    uint8
}

func decodePHC(encoded string) (phcParams, []byte, []byte, error) {
	//	check if the encoded password has the correct prefix
	if !strings.HasPrefix(encoded, phcPrefix) {
		return phcParams{}, nil, nil, errors.New("invalid phc prefix")
	}
	rest := strings.TrimPrefix(encoded, phcPrefix)

	parts := strings.Split(rest, "$")
	if len(parts) != 3 {
		return phcParams{}, nil, nil, errors.New("invalid phc format")
	}

	var memoryKiB, iterations uint32
	var threads uint64
	_, err := fmt.Sscanf(parts[0], "%d,t=%d,p=%d", &memoryKiB, &iterations, &threads)
	if err != nil {
		return phcParams{}, nil, nil, fmt.Errorf("invalid phc params: %w", err)
	}
	if threads == 0 || threads > 255 {
		return phcParams{}, nil, nil, errors.New("invalid thread count")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return phcParams{}, nil, nil, fmt.Errorf("invalid salt: %w", err)
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return phcParams{}, nil, nil, fmt.Errorf("invalid hash: %w", err)
	}

	//	return the phc params
	return phcParams{
		memoryKiB:  memoryKiB,
		iterations: iterations,
		threads:    uint8(threads),
	}, salt, hash, nil
}

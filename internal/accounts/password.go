package accounts

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// argon2id parameters from the design (S4). They travel inside the PHC string,
// so raising them later is a rehash, not a schema change.
const (
	argonMemory      = 19 * 1024 // 19456 KiB
	argonIterations  = 2
	argonParallelism = 1
	argonSaltLength  = 16
	argonKeyLength   = 32
)

// HashPassword returns a PHC-encoded argon2id hash, e.g.
// $argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("accounts: read salt: %w", err)
	}
	key := argon2.IDKey([]byte(password), salt, argonIterations, argonMemory, argonParallelism, argonKeyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonIterations, argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// VerifyPassword reports whether password matches the PHC-encoded hash. A
// malformed hash is an error, not a silent mismatch.
func VerifyPassword(password, encoded string) (bool, error) {
	params, salt, want, err := parsePHC(encoded)
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
}

func parsePHC(encoded string) (argonParams, []byte, []byte, error) {
	parts := strings.Split(encoded, "$")
	// ["", "argon2id", "v=19", "m=...,t=...,p=...", salt, hash]
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return argonParams{}, nil, nil, errors.New("accounts: not an argon2id PHC hash")
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return argonParams{}, nil, nil, errors.New("accounts: unsupported argon2 version")
	}
	var p argonParams
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism); err != nil {
		return argonParams{}, nil, nil, errors.New("accounts: malformed argon2 parameters")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return argonParams{}, nil, nil, fmt.Errorf("accounts: decode salt: %w", err)
	}
	key, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return argonParams{}, nil, nil, fmt.Errorf("accounts: decode hash: %w", err)
	}
	if len(key) == 0 {
		return argonParams{}, nil, nil, errors.New("accounts: empty hash")
	}
	return p, salt, key, nil
}

// dummyHash is a real argon2id hash of an unguessable value. Verifying against
// it when the account does not exist keeps a refused sign-in about as costly as
// a wrong password, so timing does not reveal whether an account exists (S3).
var dummyHash = mustDummyHash()

func mustDummyHash() string {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		panic("accounts: cannot read random bytes for the dummy hash: " + err.Error())
	}
	hash, err := HashPassword(base64.RawStdEncoding.EncodeToString(random))
	if err != nil {
		panic("accounts: cannot build the dummy hash: " + err.Error())
	}
	return hash
}

// VerifyDummy runs a verification against the dummy hash and always fails. It
// exists only to equalize the cost of a refused sign-in for an unknown
// identifier.
func VerifyDummy(password string) {
	_, _ = VerifyPassword(password, dummyHash)
}

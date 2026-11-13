package auth

import "golang.org/x/crypto/bcrypt"

// Indirection kept in one file so a test can count bcrypt calls.
var (
	bcryptGenerate = bcrypt.GenerateFromPassword
	bcryptCompare  = bcrypt.CompareHashAndPassword
)

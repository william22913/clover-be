package store

import "errors"

// ErrNotFound is returned when a row does not exist or is not visible.
var ErrNotFound = errors.New("not found")

// ErrEmailTaken is returned when the unique lower(email) index rejects a
// sign-up. It is the index, not a pre-check, that decides this.
var ErrEmailTaken = errors.New("email already taken")

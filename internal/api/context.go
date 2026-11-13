package api

import (
	"context"

	"github.com/william22913/clover-be/internal/store"
)

type ctxKey int

const (
	requestInfoKey ctxKey = iota
	sessionUserKey
)

// requestInfo is shared by the middleware through a pointer so the session
// layer can record the user id before the log line is written.
type requestInfo struct {
	id     string
	userID int64
}

func requestIDFrom(ctx context.Context) string {
	if info, ok := ctx.Value(requestInfoKey).(*requestInfo); ok {
		return info.id
	}
	return ""
}

func userFrom(ctx context.Context) (store.SessionUser, bool) {
	u, ok := ctx.Value(sessionUserKey).(store.SessionUser)
	return u, ok
}

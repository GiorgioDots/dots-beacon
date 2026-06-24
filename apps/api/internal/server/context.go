package server

import "context"

type contextKey string

const userIDKey contextKey = "user_id"

// WithUserID returns a copy of ctx carrying the authenticated user id.
func WithUserID(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

// UserIDFrom extracts the authenticated user id placed on the context by the
// auth middleware. The bool is false when the request was not authenticated.
func UserIDFrom(ctx context.Context) (string, bool) {
	s, ok := ctx.Value(userIDKey).(string)
	return s, ok
}

package auth

import "context"

// Repository loads and persists auth tokens and credentials.
type Repository interface {
	FindActiveSession(ctx context.Context, tokenHash string) (*Session, error)
}

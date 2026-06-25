package auth

// Session represents an authenticated refresh-token session (stub).
type Session struct {
	ID        uint
	UserID    uint
	Revoked   bool
	ExpiresAt int64
}

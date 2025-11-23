package auth

import (
	"context"
	"time"
)

// Session represents a user's authenticated state.
type Session struct {
	ID         string // Public session identifier
	ClientID   string // The user's identity, not compulsory, bucKey
	SessionKey []byte
	CreatedAt  int64 // When session was created, unix seconds
	ExpiresAt  int64 // When it should expire, unix seconds
}

// Authenticator defines the common interface for authentication flows.
type Authenticator interface {
	// NewSession creates a new entry for a session with a TTL.
	// Proper cleanup of session on any failure in guaranteed internally.
	NewSession(ctx context.Context, id, clientId string, key []byte, ttl time.Duration) error
	// GetSession returns the Session if it exists.
	GetSession(ctx context.Context, id string) (*Session, error)

	// Session Management
	ValidateSession(ctx context.Context, sessionID string) (*Session, error)
	RevokeSession(ctx context.Context, sessionID string) error
	// CleanupExpiredSessions(ctx context.Context) error
}

package redisstore

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/go-redis/redis/v8"
)

// SessionData represents what is stored in Redis for an authenticated session.
// It contains the user ID and optionally the provider tokens if needed for
// further API calls. Additional fields can be added over time.
type SessionData struct {
    UserID    string            `json:"user_id"`
    Provider  string            `json:"provider"`
    TokenData map[string]string `json:"token_data"`
}

// SessionStore encapsulates storage of session data in Redis. Keys are
// arbitrary strings (session IDs) and values are marshalled as JSON. The
// expiration is set on creation; retrieval does not alter TTL.
type SessionStore interface {
    Set(ctx context.Context, sessionID string, data SessionData, ttl time.Duration) error
    Get(ctx context.Context, sessionID string) (*SessionData, error)
    Delete(ctx context.Context, sessionID string) error
}

// redisSessionStore is a concrete implementation backed by go-redis client.
type redisSessionStore struct {
    client *redis.Client
}

// NewSessionStore constructs a new redis-backed SessionStore.
func NewSessionStore(client *redis.Client) SessionStore {
    return &redisSessionStore{client: client}
}

// Set stores session data with a TTL. It marshals the data to JSON.
func (s *redisSessionStore) Set(ctx context.Context, sessionID string, data SessionData, ttl time.Duration) error {
    b, err := json.Marshal(data)
    if err != nil {
        return fmt.Errorf("marshal session data: %w", err)
    }
    return s.client.Set(ctx, sessionID, b, ttl).Err()
}

// Get retrieves session data for the given sessionID. If the key does not
// exist, (nil, nil) is returned. Errors from Redis are propagated.
func (s *redisSessionStore) Get(ctx context.Context, sessionID string) (*SessionData, error) {
    res, err := s.client.Get(ctx, sessionID).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, nil
        }
        return nil, err
    }
    var data SessionData
    if err := json.Unmarshal([]byte(res), &data); err != nil {
        return nil, fmt.Errorf("unmarshal session data: %w", err)
    }
    return &data, nil
}

// Delete removes a session from Redis. It is idempotent.
func (s *redisSessionStore) Delete(ctx context.Context, sessionID string) error {
    return s.client.Del(ctx, sessionID).Err()
}
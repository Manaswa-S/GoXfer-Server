package redisauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"goxfer/server/internal/auth"
	"time"

	"github.com/redis/go-redis/v9"
)

// TODO: has alot of bugs

type RedisAuth struct {
	redis    *redis.Client
	redisKey string
}

func NewRedisAuth(redis *redis.Client) *RedisAuth {
	return &RedisAuth{
		redis:    redis,
		redisKey: "session-id-",
	}
}

var ErrSessionExpired = errors.New("SESSION_EXPIRED")

func (r *RedisAuth) buildKey(id string) string {
	return fmt.Sprintf("%s%s", r.redisKey, id)
}

func (r *RedisAuth) NewSession(ctx context.Context, id, clientId string, key []byte, ttl time.Duration) error {
	session := &auth.Session{
		ID:         id,
		ClientID:   clientId,
		SessionKey: key,
		CreatedAt:  time.Now().Unix(),
		ExpiresAt:  int64(ttl.Seconds()),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %v", err)
	}
	redisKey := r.buildKey(id)

	err = r.redis.SetEx(ctx, redisKey, data, ttl).Err()
	if err != nil {
		r.redis.Del(ctx, redisKey)
		return fmt.Errorf("failed to set key: %v", err)
	}

	return nil
}

func (r *RedisAuth) GetSession(ctx context.Context, id string) (*auth.Session, error) {
	redisKey := r.buildKey(id)
	data, err := r.redis.Get(ctx, redisKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch key from redis: %v", err)
	}

	session := new(auth.Session)
	err = json.Unmarshal([]byte(data), session)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %v", err)
	}

	return session, nil
}

func (r *RedisAuth) ValidateSession(ctx context.Context, id string) (*auth.Session, error) {
	session, err := r.GetSession(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %v", err)
	}

	// PERFORM CHECKS
	if !time.Now().After(time.Unix(session.ExpiresAt, 0)) {
		err = r.redis.Del(ctx, r.buildKey(id)).Err()
		if err != nil {
			return nil, fmt.Errorf("failed to del: %v", err)
		}
		return nil, ErrSessionExpired
	}

	return session, nil
}

func (r *RedisAuth) RevokeSession(ctx context.Context, id string) error {
	redisKey := r.buildKey(id)
	err := r.redis.Del(ctx, redisKey).Err()
	if err != nil {
		return fmt.Errorf("failed to del: %v", err)
	}

	return nil
}

// TODO: add sliding expiration, active sessions stay alive.

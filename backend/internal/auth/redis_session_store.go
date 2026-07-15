package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "session:"

type RedisSessionStore struct {
	client *redisclient.Client
}

func NewRedisSessionStore(client *redisclient.Client) *RedisSessionStore {
	return &RedisSessionStore{client: client}
}

func (store *RedisSessionStore) Create(
	ctx context.Context,
	userID string,
	ttl time.Duration,
) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	if err := store.client.Set(ctx, sessionKeyPrefix+token, userID, ttl).Err(); err != nil {
		return "", fmt.Errorf("set: %w", err)
	}

	return token, nil
}

func (store *RedisSessionStore) UserID(ctx context.Context, token string) (string, error) {
	userID, err := store.client.Get(ctx, sessionKeyPrefix+token).Result()
	if errors.Is(err, redisclient.Nil) {
		return "", ErrSessionNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get: %w", err)
	}

	return userID, nil
}

func (store *RedisSessionStore) Delete(ctx context.Context, token string) error {
	if err := store.client.Del(ctx, sessionKeyPrefix+token).Err(); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

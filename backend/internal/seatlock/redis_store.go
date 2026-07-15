package seatlock

import (
	"context"
	"errors"
	"fmt"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

const seatLockKeyPrefix = "seat_lock:"

var releaseScript = redisclient.NewScript(`
local owner = redis.call("GET", KEYS[1])
if not owner then
  return 0
end
if owner ~= ARGV[1] then
  return -1
end
return redis.call("DEL", KEYS[1])
`)

type RedisStore struct {
	client *redisclient.Client
}

func NewRedisStore(client *redisclient.Client) *RedisStore {
	return &RedisStore{client: client}
}

func (store *RedisStore) Acquire(
	ctx context.Context,
	screeningID string,
	seatID string,
	userID string,
	ttl time.Duration,
) (Lock, error) {
	key := seatLockKey(screeningID, seatID)

	for attempt := 0; attempt < 2; attempt++ {
		acquired, err := store.client.SetNX(ctx, key, userID, ttl).Result()
		if err != nil {
			return Lock{}, fmt.Errorf("set if absent: %w", err)
		}
		if acquired {
			return Lock{
				ScreeningID: screeningID,
				SeatID:      seatID,
				UserID:      userID,
				ExpiresAt:   time.Now().UTC().Add(ttl),
			}, nil
		}

		owner, err := store.client.Get(ctx, key).Result()
		if errors.Is(err, redisclient.Nil) {
			continue
		}
		if err != nil {
			return Lock{}, fmt.Errorf("get owner: %w", err)
		}
		if owner != userID {
			return Lock{}, ErrAlreadyLocked
		}

		remaining, err := store.client.PTTL(ctx, key).Result()
		if err != nil {
			return Lock{}, fmt.Errorf("get expiry: %w", err)
		}
		if remaining <= 0 {
			continue
		}

		return Lock{
			ScreeningID: screeningID,
			SeatID:      seatID,
			UserID:      userID,
			ExpiresAt:   time.Now().UTC().Add(remaining),
		}, nil
	}

	return Lock{}, ErrAlreadyLocked
}

func (store *RedisStore) Current(
	ctx context.Context,
	screeningID string,
	seatIDs []string,
) (map[string]Lock, error) {
	locks := make(map[string]Lock)
	if len(seatIDs) == 0 {
		return locks, nil
	}

	keys := make([]string, 0, len(seatIDs))
	for _, seatID := range seatIDs {
		keys = append(keys, seatLockKey(screeningID, seatID))
	}

	owners, err := store.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("get owners: %w", err)
	}

	pipeline := store.client.Pipeline()
	ttlCommands := make(map[int]*redisclient.DurationCmd)
	for index, owner := range owners {
		if owner != nil {
			ttlCommands[index] = pipeline.PTTL(ctx, keys[index])
		}
	}
	if len(ttlCommands) > 0 {
		if _, err := pipeline.Exec(ctx); err != nil {
			return nil, fmt.Errorf("get expiries: %w", err)
		}
	}

	now := time.Now().UTC()
	for index, command := range ttlCommands {
		remaining := command.Val()
		owner, ok := owners[index].(string)
		if !ok || remaining <= 0 {
			continue
		}

		seatID := seatIDs[index]
		locks[seatID] = Lock{
			ScreeningID: screeningID,
			SeatID:      seatID,
			UserID:      owner,
			ExpiresAt:   now.Add(remaining),
		}
	}

	return locks, nil
}

func (store *RedisStore) Release(
	ctx context.Context,
	screeningID string,
	seatID string,
	userID string,
) error {
	result, err := releaseScript.Run(
		ctx,
		store.client,
		[]string{seatLockKey(screeningID, seatID)},
		userID,
	).Int64()
	if err != nil {
		return fmt.Errorf("compare and delete: %w", err)
	}
	if result == -1 {
		return ErrLockNotOwned
	}

	return nil
}

func seatLockKey(screeningID, seatID string) string {
	return seatLockKeyPrefix + screeningID + ":" + seatID
}

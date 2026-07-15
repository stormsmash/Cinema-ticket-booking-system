package seatlock

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
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

var claimScript = redisclient.NewScript(`
local owner = redis.call("GET", KEYS[1])
if not owner then
  return {0, 0}
end
if owner ~= ARGV[1] then
  return {-1, 0}
end
local remaining = redis.call("PTTL", KEYS[1])
if remaining <= 0 then
  return {0, 0}
end
redis.call("SET", KEYS[1], ARGV[2], "PX", ARGV[3], "XX")
return {1, remaining}
`)

var restoreClaimScript = redisclient.NewScript(`
local owner = redis.call("GET", KEYS[1])
if owner ~= ARGV[1] then
  return 0
end
local remaining = tonumber(ARGV[3])
if remaining <= 0 then
  return redis.call("DEL", KEYS[1])
end
redis.call("SET", KEYS[1], ARGV[2], "PX", remaining, "XX")
return 1
`)

var commitClaimScript = redisclient.NewScript(`
local owner = redis.call("GET", KEYS[1])
if owner ~= ARGV[1] then
  return 0
end
return redis.call("DEL", KEYS[1])
`)

type Claim struct {
	ScreeningID       string
	SeatID            string
	UserID            string
	Token             string
	OriginalExpiresAt time.Time
}

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
			UserID:      lockOwner(owner),
			ExpiresAt:   now.Add(remaining),
		}
	}

	return locks, nil
}

func (store *RedisStore) Claim(
	ctx context.Context,
	screeningID string,
	seatID string,
	userID string,
	claimTTL time.Duration,
) (Claim, error) {
	token, err := randomClaimToken()
	if err != nil {
		return Claim{}, err
	}

	claimValue := seatLockClaimValue(userID, token)
	result, err := claimScript.Run(
		ctx,
		store.client,
		[]string{seatLockKey(screeningID, seatID)},
		userID,
		claimValue,
		claimTTL.Milliseconds(),
	).Int64Slice()
	if err != nil {
		return Claim{}, fmt.Errorf("claim seat lock: %w", err)
	}
	if len(result) != 2 {
		return Claim{}, fmt.Errorf("claim seat lock: unexpected Redis response")
	}
	if result[0] == 0 {
		return Claim{}, ErrLockNotFound
	}
	if result[0] == -1 {
		return Claim{}, ErrLockNotOwned
	}

	return Claim{
		ScreeningID:       screeningID,
		SeatID:            seatID,
		UserID:            userID,
		Token:             token,
		OriginalExpiresAt: time.Now().UTC().Add(time.Duration(result[1]) * time.Millisecond),
	}, nil
}

func (store *RedisStore) RestoreClaim(ctx context.Context, claim Claim) error {
	remaining := time.Until(claim.OriginalExpiresAt)
	if _, err := restoreClaimScript.Run(
		ctx,
		store.client,
		[]string{seatLockKey(claim.ScreeningID, claim.SeatID)},
		seatLockClaimValue(claim.UserID, claim.Token),
		claim.UserID,
		remaining.Milliseconds(),
	).Result(); err != nil {
		return fmt.Errorf("restore seat lock claim: %w", err)
	}

	return nil
}

func (store *RedisStore) CommitClaim(ctx context.Context, claim Claim) error {
	if _, err := commitClaimScript.Run(
		ctx,
		store.client,
		[]string{seatLockKey(claim.ScreeningID, claim.SeatID)},
		seatLockClaimValue(claim.UserID, claim.Token),
	).Result(); err != nil {
		return fmt.Errorf("commit seat lock claim: %w", err)
	}

	return nil
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
	if result == 0 {
		return ErrLockNotFound
	}

	return nil
}

func seatLockKey(screeningID, seatID string) string {
	return seatLockKeyPrefix + screeningID + ":" + seatID
}

func seatLockClaimValue(userID, token string) string {
	return "booking_claim:" + userID + ":" + token
}

func lockOwner(value string) string {
	if !strings.HasPrefix(value, "booking_claim:") {
		return value
	}

	value = strings.TrimPrefix(value, "booking_claim:")
	owner, _, found := strings.Cut(value, ":")
	if !found {
		return ""
	}
	return owner
}

func randomClaimToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate claim token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

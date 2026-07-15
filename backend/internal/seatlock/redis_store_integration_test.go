package seatlock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

func TestRedisStoreAllowsOnlyOneWinnerForConcurrentSeatLock(t *testing.T) {
	address := os.Getenv("REDIS_TEST_ADDRESS")
	if address == "" {
		t.Skip("set REDIS_TEST_ADDRESS to run the Redis concurrency test")
	}

	client := redisclient.NewClient(&redisclient.Options{Addr: address, DB: 15})
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("connect to test Redis: %v", err)
	}

	store := NewRedisStore(client)
	screeningID := fmt.Sprintf("concurrency-%d", time.Now().UnixNano())
	const (
		seatID  = "A1"
		workers = 32
	)
	key := seatLockKey(screeningID, seatID)
	t.Cleanup(func() { _ = client.Del(context.Background(), key).Err() })

	type result struct {
		userID string
		err    error
	}
	results := make(chan result, workers)
	start := make(chan struct{})
	var waitGroup sync.WaitGroup
	for index := 0; index < workers; index++ {
		userID := fmt.Sprintf("user-%02d", index)
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start
			_, err := store.Acquire(context.Background(), screeningID, seatID, userID, time.Minute)
			results <- result{userID: userID, err: err}
		}()
	}

	close(start)
	waitGroup.Wait()
	close(results)

	winner := ""
	lockedCount := 0
	for item := range results {
		switch {
		case item.err == nil:
			if winner != "" {
				t.Fatalf("more than one user acquired the same seat: %s and %s", winner, item.userID)
			}
			winner = item.userID
		case errors.Is(item.err, ErrAlreadyLocked):
			lockedCount++
		default:
			t.Fatalf("unexpected acquire error for %s: %v", item.userID, item.err)
		}
	}

	if winner == "" {
		t.Fatal("expected one user to acquire the seat")
	}
	if lockedCount != workers-1 {
		t.Fatalf("expected %d rejected users, got %d", workers-1, lockedCount)
	}
	owner, err := client.Get(context.Background(), key).Result()
	if err != nil {
		t.Fatalf("read winning lock owner: %v", err)
	}
	if owner != winner {
		t.Fatalf("expected Redis owner %s, got %s", winner, owner)
	}
}

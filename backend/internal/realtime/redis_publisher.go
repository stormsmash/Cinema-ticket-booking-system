package realtime

import (
	"context"
	"encoding/json"
	"fmt"

	redisclient "github.com/redis/go-redis/v9"
)

const seatEventChannel = "cinema:seat-events:v1"

type RedisPublisher struct {
	client *redisclient.Client
}

func NewRedisPublisher(client *redisclient.Client) *RedisPublisher {
	return &RedisPublisher{client: client}
}

func (publisher *RedisPublisher) Publish(ctx context.Context, event SeatEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode seat event: %w", err)
	}
	if err := publisher.client.Publish(ctx, seatEventChannel, payload).Err(); err != nil {
		return fmt.Errorf("publish seat event: %w", err)
	}

	return nil
}

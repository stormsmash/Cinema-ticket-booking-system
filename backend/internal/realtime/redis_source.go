package realtime

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	redisclient "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const seatLockKeyPrefix = "seat_lock:"

const (
	initialReconnectDelay = time.Second
	maximumReconnectDelay = 10 * time.Second
)

type RedisSeatEventSource struct {
	client *redisclient.Client
	db     int
	now    func() time.Time
}

func NewRedisSeatEventSource(client *redisclient.Client, db int) *RedisSeatEventSource {
	return &RedisSeatEventSource{
		client: client,
		db:     db,
		now:    func() time.Time { return time.Now().UTC() },
	}
}

func (source *RedisSeatEventSource) Run(
	ctx context.Context,
	publish func(SeatEvent),
) error {
	reconnectDelay := initialReconnectDelay
	for {
		err := source.subscribe(ctx, publish)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			log.Printf("Redis seat event subscription interrupted: %v", err)
		}

		timer := time.NewTimer(reconnectDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}

		reconnectDelay = min(reconnectDelay*2, maximumReconnectDelay)
	}
}

func (source *RedisSeatEventSource) subscribe(
	ctx context.Context,
	publish func(SeatEvent),
) error {
	channels := []string{
		source.eventChannel("set"),
		source.eventChannel("del"),
		source.eventChannel("expired"),
	}
	pubsub := source.client.Subscribe(ctx, channels...)
	defer pubsub.Close()

	if _, err := pubsub.Receive(ctx); err != nil {
		return fmt.Errorf("subscribe to Redis key events: %w", err)
	}

	messages := pubsub.Channel(redisclient.WithChannelSize(64))
	for {
		select {
		case <-ctx.Done():
			return nil
		case message, open := <-messages:
			if !open {
				if ctx.Err() != nil {
					return nil
				}
				return errors.New("Redis key event subscription closed")
			}

			event, ok, err := source.toSeatEvent(ctx, message.Channel, message.Payload)
			if err != nil {
				log.Printf("process Redis seat event: %v", err)
				continue
			}
			if ok {
				publish(event)
			}
		}
	}
}

func (source *RedisSeatEventSource) toSeatEvent(
	ctx context.Context,
	channel string,
	key string,
) (SeatEvent, bool, error) {
	screeningID, seatID, ok := parseSeatLockKey(key)
	if !ok {
		return SeatEvent{}, false, nil
	}

	eventName := strings.TrimPrefix(channel, source.eventChannel(""))
	now := source.now()
	event := SeatEvent{
		Version:     EventVersion,
		ScreeningID: screeningID,
		SeatID:      seatID,
		OccurredAt:  now,
	}

	switch eventName {
	case "set":
		remaining, err := source.client.PTTL(ctx, key).Result()
		if err != nil {
			return SeatEvent{}, false, fmt.Errorf("read seat lock expiry: %w", err)
		}
		if remaining <= 0 {
			return SeatEvent{}, false, nil
		}

		expiresAt := now.Add(remaining)
		event.Type = SeatLocked
		event.Status = "LOCKED"
		event.ExpiresAt = &expiresAt
		return event, true, nil
	case "del", "expired":
		exists, err := source.client.Exists(ctx, key).Result()
		if err != nil {
			return SeatEvent{}, false, fmt.Errorf("check current seat lock: %w", err)
		}
		if exists > 0 {
			return SeatEvent{}, false, nil
		}

		if eventName == "del" {
			event.Type = SeatReleased
		} else {
			event.Type = SeatExpired
		}
		event.Status = "AVAILABLE"
		return event, true, nil
	default:
		return SeatEvent{}, false, nil
	}
}

func (source *RedisSeatEventSource) eventChannel(eventName string) string {
	return fmt.Sprintf("__keyevent@%d__:%s", source.db, eventName)
}

func parseSeatLockKey(key string) (string, string, bool) {
	value, found := strings.CutPrefix(key, seatLockKeyPrefix)
	if !found {
		return "", "", false
	}

	screeningID, seatID, found := strings.Cut(value, ":")
	if !found || screeningID == "" || seatID == "" || strings.Contains(seatID, ":") {
		return "", "", false
	}
	if _, err := bson.ObjectIDFromHex(screeningID); err != nil {
		return "", "", false
	}

	return screeningID, seatID, true
}

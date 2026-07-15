package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID            bson.ObjectID `bson:"_id,omitempty"`
	GoogleSubject string        `bson:"google_subject"`
	Email         string        `bson:"email"`
	Name          string        `bson:"name"`
	AvatarURL     string        `bson:"avatar_url,omitempty"`
	CreatedAt     time.Time     `bson:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at"`
}

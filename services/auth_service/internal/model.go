package auth

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	DeviceTokens []string `json:"device_tokens,omitempty"`
}

type UserMongo struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	Password     string             `bson:"password" json:"password"`
	DeviceTokens []string           `bson:"device_tokens,omitempty" json:"device_tokens,omitempty"`
}

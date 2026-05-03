package auth

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID           string   `json:"id"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	DisplayName  string   `json:"display_name,omitempty"`
	DeviceTokens []string `json:"device_tokens,omitempty"`
}

type UserMongo struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	Password     string             `bson:"password" json:"password"`
	DisplayName  string             `bson:"display_name,omitempty" json:"display_name,omitempty"`
	DeviceTokens []string           `bson:"device_tokens,omitempty" json:"device_tokens,omitempty"`
}

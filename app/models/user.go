package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID                   bson.ObjectID `json:"_id,omitempty"  bson:"_id,omitempty"`
	UserID               string        `json:"user_id,omitempty" bson:"user_id"`
	FirstName            string        `json:"first_name,omitempty" bson:"first_name,omitempty" validate:"required"`
	LastName             string        `json:"last_name,omitempty" bson:"last_name,omitempty" validate:"required"`
	Email                string        `json:"email,omitempty" bson:"email,omitempty" validate:"required,email"`
	Password             string        `json:"password,omitempty" bson:"password,omitempty" validate:"required"`
	Role                 string        `json:"role,,omitempty" bson:"role,,omitempty"`
	CreatedAt            time.Time     `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt            time.Time     `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	PasswordResetToken   string        `json:"password_reset_token,omitempty" bson:"password_reset_token,omitempty"`
	PasswordTokenExpired time.Time     `json:"password_token_expired,omitempty" bson:"password_token_expired,omitempty"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Token     string `json:"token"`
}

type PasswordUpdate struct {
	NewPassword     string `json:"new_password" validate:"required,max=16,min=6"`
	CurrentPassword string `json:"current_password" validate:"required"`
}

type Category struct {
	ID       bson.ObjectID `json:"_id,omitempty"  bson:"_id,omitempty"`
	Category string        `json:"category,omitempty" bson:"category,omitempty"`
}

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Image represents an image stored in S3 with metadata in MongoDB
type Image struct {
	ID         bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Category   string        `json:"category" bson:"category"` // e.g., "anime", "games"
	FileName   string        `json:"file_name" bson:"file_name"`
	S3Key      string        `json:"s3_key" bson:"s3_key"` // Full S3 path: anime/image.jpg
	S3URL      string        `json:"s3_url" bson:"s3_url"` // Full accessible URL
	UploadedAt time.Time     `json:"uploaded_at" bson:"uploaded_at"`
}

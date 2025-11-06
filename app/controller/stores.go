package controller

import (
	"context"
	"fmt"
	"ginmongo/database"
	"ginmongo/models"
	"ginmongo/utils"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ImageResponse struct {
	Category   string    `json:"category"`
	FileName   string    `json:"file_name"`
	S3Key      string    `json:"s3_key"`
	S3URL      string    `json:"s3_url"`
	SignedURL  string    `json:"signed_url"`
	UploadedAt time.Time `json:"uploaded_at"`
}

var s3Client *s3.Client
var presignClient *s3.PresignClient

func InitS3Client() {
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Println(err)
		return
	}
	s3Client = s3.NewFromConfig(cfg)
	presignClient = s3.NewPresignClient(s3Client)
}

func UploadImage(c *gin.Context) {

	userRole, exists := c.Get("role")
	log.Println(userRole.(string))
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found"})
		return
	}

	ok, err := utils.Authroizeuser(userRole.(string), "admin")

	if err != nil || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User unauthorized"})
		return
	}
	bucketName := os.Getenv("BUCKET_NAME")
	region := os.Getenv("AWS_REGION")
	category := c.Param("category")

	file, err := c.FormFile("image")

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "No image file provided"})
		return
	}
	fileContent, err := file.Open()

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	defer fileContent.Close()

	s3Key := fmt.Sprintf("%s/%s", category, file.Filename)

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Key:    &s3Key,
		Bucket: &bucketName,
		Body:   fileContent,
	})

	if err != nil {
		log.Println(err)
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Error uploading images"})
		return
	}

	s3Url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, s3Key)

	ImageDoc := &models.Image{
		Category:   category,
		FileName:   file.Filename,
		S3Key:      s3Key,
		S3URL:      s3Url,
		UploadedAt: time.Now(),
	}

	collection := database.Client.Database("imagestore").Collection("images")

	result, err := collection.InsertOne(context.TODO(), ImageDoc)

	if err != nil {
		log.Println("Mongo Insert :", err)
		return
	}
	c.IndentedJSON(http.StatusOK, result)
}

func GetImagesByCategory(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	category := c.Param("category")
	bucketName := os.Getenv("BUCKET_NAME")

	collection := database.Client.Database("imagestore").Collection("images")

	// --- Pagination parameters ---
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 6
	}
	skip := (page - 1) * limit

	// Create filter for the category
	filter := bson.M{"category": category}

	// Count documents MATCHING THE FILTER (not all documents)
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Println("Error counting documents:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Find options with pagination and sorting
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "uploadedAt", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting images"})
		return
	}
	defer cursor.Close(ctx)

	var images []models.Image
	if err = cursor.All(ctx, &images); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing images"})
		return
	}

	// Generate pre-signed URLs for each image
	responseImages := make([]ImageResponse, 0, len(images))
	for _, img := range images {
		// Generate pre-signed URL valid for 10 minutes
		request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(img.S3Key),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(10 * time.Minute)
		})

		signedURL := img.S3URL
		if err == nil {
			signedURL = request.URL
		} else {
			log.Println("Error generating pre-signed URL:", err)
		}

		responseImages = append(responseImages, ImageResponse{
			Category:   img.Category,
			FileName:   img.FileName,
			S3Key:      img.S3Key,
			S3URL:      img.S3URL,
			SignedURL:  signedURL,
			UploadedAt: img.UploadedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"images":     responseImages,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": int(math.Ceil(float64(total) / float64(limit))),
	})
}

func GetAllImages(c *gin.Context) {

	var images []models.Image
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	bucketName := os.Getenv("BUCKET_NAME")

	collection := database.Client.Database("imagestore").Collection("images")
	// --- Pagination parameters ---
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 6
	}
	skip := (page - 1) * limit
	total, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Println("Error counting documents:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "uploadedAt", Value: -1}})

	curson, err := collection.Find(ctx, bson.M{}, findOptions)

	defer curson.Close(ctx)

	if err = curson.All(ctx, &images); err != nil {
		log.Println(err)
		return
	}

	var responseImages []ImageResponse
	for _, img := range images {
		// Generate pre-signed URL valid for 1 hour
		request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(img.S3Key),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = time.Duration(60 * time.Minute)
		})

		signedURL := img.S3URL
		if err == nil {
			signedURL = request.URL
		} else {
			log.Println("Error generating pre-signed URL:", err)
		}

		responseImages = append(responseImages, ImageResponse{
			Category:   img.Category,
			FileName:   img.FileName,
			S3Key:      img.S3Key,
			S3URL:      img.S3URL,
			SignedURL:  signedURL,
			UploadedAt: img.UploadedAt,
		})
	}

	//
	c.JSON(http.StatusOK, gin.H{
		"images":     responseImages,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": int(math.Ceil(float64(total) / float64(limit))),
	})
}

func GetImagesByName(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	name := c.Query("name")

	log.Println(name)
	bucketName := os.Getenv("BUCKET_NAME")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' query parameter"})
		return
	}

	collection := database.Client.Database("imagestore").Collection("images")
	// --- Pagination parameters ---
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 6
	}
	skip := (page - 1) * limit
	// Case-insensitive regex search
	// Build regex: "god war" -> ".*god[-_ ]*war.*"
	pattern := strings.TrimSpace(name)
	pattern = regexp.MustCompile(`\s+`).ReplaceAllString(pattern, "[-_ ]*")
	pattern = fmt.Sprintf(".*%s.*", pattern)
	filter := bson.M{
		"file_name": bson.M{
			"$regex":   pattern,
			"$options": "i", // case-insensitive
		},
	}
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Println("Error counting documents:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	findOptions := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "uploadedAt", Value: -1}})

	//opts := options.Find().SetSort(bson.D{{Key: "uploaded_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Println("Mongo find error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error searching images"})
		return
	}
	defer cursor.Close(ctx)

	var images []models.Image
	if err = cursor.All(ctx, &images); err != nil {
		log.Println("Mongo cursor error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing results"})
		return
	}

	var responseImages []ImageResponse
	for _, img := range images {
		request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(img.S3Key),
		}, func(opts *s3.PresignOptions) {
			opts.Expires = 10 * time.Minute
		})

		signedURL := img.S3URL
		if err == nil {
			signedURL = request.URL
		} else {
			log.Println("Error generating pre-signed URL:", err)
		}

		responseImages = append(responseImages, ImageResponse{
			Category:   img.Category,
			FileName:   img.FileName,
			S3Key:      img.S3Key,
			S3URL:      img.S3URL,
			SignedURL:  signedURL,
			UploadedAt: img.UploadedAt,
		})
	}

	// c.JSON(http.StatusOK, gin.H{
	// 	"search": name,
	// 	"images": responseImages,
	// 	"total":  len(responseImages),
	// })
	c.JSON(http.StatusOK, gin.H{
		"images":     responseImages,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": int(math.Ceil(float64(total) / float64(limit))),
	})
}

func CreateCategory(c *gin.Context) {
	userRole, exists := c.Get("role")
	log.Println(userRole.(string))
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}

	ok, err := utils.Authroizeuser(userRole.(string), "admin")

	if err != nil || !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var category models.Category

	if err := c.ShouldBind(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "INvalid payload"})
		return
	}

	collection := database.Client.Database("imagestore").Collection("categories")

	_, err = collection.InsertOne(ctx, bson.M{"category": category.Category})
	if err != nil {
		log.Println(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   true,
		"category": category.Category,
	})
}

func GetCategories(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var category []models.Category

	collection := database.Client.Database("imagestore").Collection("categories")

	total, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Println("Error counting documents:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		log.Println(err)
		return
	}
	if err = cursor.All(ctx, &category); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing images"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Category": category, "total": total})
}

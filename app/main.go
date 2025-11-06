package main

import (
	"ginmongo/controller"
	"ginmongo/database"
	"ginmongo/route"
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	controller.InitS3Client()

	// Initialize MongoDB connection
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	router := gin.Default()
	// router.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"http://front:3000", "http://localhost:3001"}, // Allows all localhost ports
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Authorization"},
	// 	ExposeHeaders:    []string{"Authorization"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// }))
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:") ||
				strings.HasPrefix(origin, "http://")
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// rateLimit := middlewares.NewRateLimiter(1000, time.Minute)
	// router.Use(rateLimit.Middleware())
	route.Protected(router)
	route.Unprotected(router)

	router.Run(":8007")
}

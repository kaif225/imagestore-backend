package middlewares

import (
	"fmt"
	"ginmongo/utils"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWT() gin.HandlerFunc {
	// return func(c *gin.Context) {
	// 	tokenCookie, err := c.Request.Cookie("Bearer")
	// 	if err != nil {
	// 		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Authorization cookie is missing"})
	// 		return
	// 	}
	// 	// authHeader := c.GetHeader("Authorization")
	// 	// if authHeader == "" {
	// 	// 	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Authorization header missing"})
	// 	// 	return
	// 	// }

	// 	// // Expecting format: "Bearer <token>"
	// 	// parts := strings.SplitN(authHeader, " ", 2)
	// 	// if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
	// 	// 	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Invalid Authorization header format"})
	// 	// 	return
	// 	// }

	// 	// tokenString := parts[1]
	return func(c *gin.Context) {
		var tokenString string

		// Try to get token from Authorization header first (for Postman/API clients)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// Expecting format: "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				tokenString = parts[1]
			}
		}

		// If not in header, try cookie (for browser)
		if tokenString == "" {
			tokenCookie, err := c.Request.Cookie("Bearer")
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Authorization token required",
				})
				return
			}
			tokenString = tokenCookie.Value
		}

		jwtSecret := os.Getenv("JWT_SECRET")
		claims := &utils.SignedDetails{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Ensure signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if claims.ExpiresAt.Time.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
			return
		}

		c.Set("role", claims.Role)
		//c.Set("userID", claims.)

		c.Next()
	}
}

package utils

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SignedDetails struct {
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	jwt.RegisteredClaims
}

func SignedToken(email, firstName, lastName, role string) (string, error) {

	secret := os.Getenv("JWT_SECRET")
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "imagesearch",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Println(err)
		return "", errors.New("Error in signing")
	}
	return signedToken, nil
}

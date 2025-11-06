package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"

	"golang.org/x/crypto/argon2"
)

func HashPass(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)

	if err != nil {
		log.Println(err)
		return "", errors.New("Unable to create salt")
	}

	Hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	HashBase64 := base64.StdEncoding.EncodeToString(Hash)

	HassPass := fmt.Sprintf("%s.%s", saltBase64, HashBase64)
	return HassPass, nil
}

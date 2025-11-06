package utils

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log"
	"strings"

	"golang.org/x/crypto/argon2"
)

func ComparePass(password, hashPassword string) error {
	parts := strings.Split(hashPassword, ".")
	if len(parts) != 2 {
		return errors.New("invalid format")
	}
	saltBase64 := parts[0]
	hashBase64 := parts[1]

	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		log.Println(err)
		return nil
	}
	hash, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		log.Println(err)
		return nil
	}
	Hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	if len(hash) != len(Hash) {
		return errors.New("Incorrect Password")
	}

	if subtle.ConstantTimeCompare(hash, Hash) != 1 {
		return errors.New("Incorrect Password")
	}
	return nil

}

package util

import (
	"golang.org/x/crypto/bcrypt"
	"proman-backend/internal/pkg/log"
)

// CheckPassword is function to compare hashed and plain text
func CheckPassword(hashed, text string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(text))
	if err != nil {
		return false
	}
	return true
}

// CryptPassword is a function to encrypt plain text to bcrypt
func CryptPassword(text string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)
	if err != nil {
		log.Panic(err)
	}
	return string(hashed)
}

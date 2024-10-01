package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func CheckPasswordHash(password, hash string) error {

	bytePassword := []byte(password)
	byteHash := []byte(hash)

	if err := bcrypt.CompareHashAndPassword(byteHash, bytePassword); err != nil {
		return fmt.Errorf("error pasword mismatch: %s", err)
	}

	return nil
}

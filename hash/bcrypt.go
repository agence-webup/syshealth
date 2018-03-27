package hash

import "golang.org/x/crypto/bcrypt"

// Create returns a hashed string of the password
func Create(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Check compares a given password with a hash
func Check(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

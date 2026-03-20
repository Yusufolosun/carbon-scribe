package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a password using bcrypt with configurable cost
func HashPassword(password string, cost ...int) (string, error) {
	hashCost := bcrypt.DefaultCost
	if len(cost) > 0 && cost[0] > 0 {
		hashCost = cost[0]
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), hashCost)
	return string(hashed), err
}

// CheckPassword verifies a password against its hash (deprecated, use VerifyPassword)
func CheckPassword(password, hashed string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(hashed, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
}

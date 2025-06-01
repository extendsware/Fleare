package utils

import "math/rand"

// generate a unique random string by using a number generator
func GenerateUniqueId(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result string
	for range length {
		result += string(charset[rand.Intn(len(charset))])
	}
	return result
}

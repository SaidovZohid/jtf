package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
)

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func main() {
	str, err := generateRandomString(16)
	if err != nil {
		log.Println("error: ", err)
	}
	fmt.Println("Random 32-byte string: ", str)
}

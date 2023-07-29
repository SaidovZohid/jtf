package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/SaidovZohid/swiftsend.it/api/handlers"
)

const (
	setKeyMsg     = "set your secret key"
	privateKeyMsg = "private key here"
)

func main() {
	secretKey := []byte(setKeyMsg)
	if string(secretKey) == setKeyMsg {
		log.Fatal("generate secret key by running this file 'go run ./pkg/random_key/main.go'")
	}

	// paste here your private key in order to encrypt it and after encyption copy/paste to .env file ENCRYPTED_PRIVATE_KEY key.
	privateKey := []byte(privateKeyMsg)
	if string(privateKey) == privateKeyMsg {
		log.Fatal("set your private key!")
	}

	// Encrypt the data
	encrypted, err := handlers.Encrypt(privateKey, secretKey)
	if err != nil {
		fmt.Println("Encryption error:", err)
		return
	}

	// Convert the encrypted data to a base64-encoded string
	encoded := base64.StdEncoding.EncodeToString(encrypted)
	fmt.Println("Encrypted data:", encoded)

	// Decrypt the data
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		fmt.Println("Decoding error:", err)
		return
	}

	decrypted, err := handlers.Decrypt(decoded, secretKey)
	if err != nil {
		fmt.Println("Decryption error:", err)
		return
	}

	fmt.Println("Decrypted data:", string(decrypted))
}

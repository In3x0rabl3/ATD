package utils

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
)

func InitializeLogger() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
}

func GenerateUniqueID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatalf("Failed to generate unique ID: %v", err)
	}
	return hex.EncodeToString(bytes)
}

package entities

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// generateID generates a unique ID for entities
func generateID() string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%d-%s", timestamp, randomHex)
}
package adr

import (
	"fmt"
	"time"
)

// GenerateID generates a unique ADR ID
func GenerateID() string {
	return fmt.Sprintf("ADR-%04d", time.Now().Unix()%10000)
}
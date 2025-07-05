package services

import (
	"fmt"
	"time"
)

func generatePeerID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

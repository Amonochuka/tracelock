package access

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func GenerateHash(userID, zoneID int, action string,
	timestamp time.Time, previousHash string) string {
	data := fmt.Sprintf("%d:%d:%s:%s:%s",
		userID, zoneID, action, timestamp.UTC().Format(time.RFC3339), previousHash)

	//hash using sha256
	hash := sha256.Sum256([]byte(data))

	//convert back to a string that is readable
	return hex.EncodeToString(hash[:])
}

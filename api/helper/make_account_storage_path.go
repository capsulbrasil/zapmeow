package helper

import (
	"fmt"
	"zapmeow/config"
)

func MakeAccountStoragePath(instanceID string) string {
	config := config.Load()
	return fmt.Sprintf("%s/instance_%s", config.StoragePath, instanceID)
}

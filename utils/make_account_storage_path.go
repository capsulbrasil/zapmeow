package utils

import (
	"fmt"
	"zapmeow/configs"
)

func MakeAccountStoragePath(instanceID string) string {
	config := configs.LoadConfigs()
	return fmt.Sprintf("%s/instance_%s", config.StoragePath, instanceID)
}

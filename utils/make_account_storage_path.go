package utils

import (
	"fmt"
	"os"
)

func MakeAccountStoragePath(instanceID string) string {
	return fmt.Sprintf("%s/instance_%s", os.Getenv("STORAGE_PATH"), instanceID)
}

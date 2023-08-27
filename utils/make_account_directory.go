package utils

import (
	"fmt"
	"os"
)

func MakeUserDirectory(user string) (string, error) {
	dir := fmt.Sprintf("%s/user_%s", os.Getenv("STORAGE_PATH"), user)
	err := os.MkdirAll(dir, 0751)
	if err != nil {
		return "", err
	}

	return dir, nil
}

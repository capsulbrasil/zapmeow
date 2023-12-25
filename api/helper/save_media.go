package helper

import (
	"fmt"
	"mime"
	"os"
)

func SaveMedia(instanceID string, fileName string, data []byte, mimetype string) (string, error) {
	dirPath := MakeAccountStoragePath(instanceID)
	err := os.MkdirAll(dirPath, 0751)
	if err != nil {
		return "", err
	}

	exts, err := mime.ExtensionsByType(mimetype)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("%s/%s%s", dirPath, fileName, exts[0])

	err = os.WriteFile(path, data, 0600)
	if err != nil {
		return "", err
	}

	return path, nil
}

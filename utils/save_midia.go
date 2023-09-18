package utils

import (
	"fmt"
	"mime"
	"os"
)

func SaveMedia(instanceID string, data []byte, fileName string, mimetype string) (string, error) {
	dirPath := MakeAccountStoragePath(instanceID)
	err := os.MkdirAll(dirPath, 0751)
	if err != nil {
		return "", err
	}

	exts, _ := mime.ExtensionsByType(mimetype)
	path := fmt.Sprintf("%s/%s%s", dirPath, fileName, exts[0])

	err = os.WriteFile(path, data, 0600)
	if err != nil {
		// fmt.Println("failed to save file", err)
		return "", err
	}

	// fmt.Println("file saved: ", path)
	return path, nil
}

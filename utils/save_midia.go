package utils

import (
	"fmt"
	"os"
)

func SaveMedia(data []byte, dir string, fileName string, ext string) (string, error) {
	path := fmt.Sprintf("%s/%s%s", dir, fileName, ext)

	err := os.WriteFile(path, data, 0600)
	if err != nil {
		fmt.Println("failed to save file", err)
		return "", err
	}

	fmt.Println("file saved: ", path)
	return path, nil
}

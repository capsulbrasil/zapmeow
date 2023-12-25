package helper

import (
	"errors"
	"strings"
)

func GetMimeTypeFromDataURI(dataURI string) (string, error) {
	components := strings.Split(dataURI, ",")
	if len(components) < 2 {
		return "", errors.New("invalid data uri")
	}

	mimeTypeComponents := strings.Split(components[0], ";")
	if len(mimeTypeComponents) < 2 {
		return "", errors.New("mime type not found")
	}

	mimeType := strings.TrimPrefix(mimeTypeComponents[0], "data:")

	if mimeType == "audio/ogg" {
		params := mimeTypeComponents[1:]
		codecsFound := false

		for _, param := range params {
			if strings.TrimSpace(param) == "codecs=opus" {
				codecsFound = true
				break
			}
		}

		if !codecsFound {
			mimeType += "; codecs=opus"
		}
	}

	return mimeType, nil
}

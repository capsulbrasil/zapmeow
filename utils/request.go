package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func Request(url string, data map[string]interface{}) error {
	body, err := json.Marshal(data)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func Request(url string, data map[string]interface{}) error {
	body, err := json.Marshal(data)

	if err != nil {
		fmt.Println("Error when serializing the map in JSON:", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}

	defer resp.Body.Close()
	return nil
}

package utils

import (
	"encoding/json"
	"io"
	"os"
)

type Proxy struct {
	Scheme string `json:"scheme"`
	Ip     string `json:"ip"`
	Port   string `json:"port"`
}

type Proxys struct {
	Proxys []Proxy `json:"proxys"`
}

func ReadProxys(path string) (*Proxys, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)

	if err != nil {
		return nil, err
	}

	var proxys Proxys
	err = json.Unmarshal(byteValue, &proxys)
	if err != nil {
		return nil, err
	}
	return &proxys, nil
}

package sts

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ApiAddress string `json:"api_address"`
	ApiKeyPath string `json:"api_key_path"`
}

type Satellite struct {
	Id int `json:"id"`
}

type GroundStation struct {
	Id int `json:"id"`
}

func GetConfig(filePath string) (Config, error) {

	var output Config

	rawContent, err := os.ReadFile(filePath)
	if err != nil {
		return output, fmt.Errorf("Error when reading the configuration file: %w", err)
	}

	err = json.Unmarshal(rawContent, &output)
	if err != nil {
		return output, fmt.Errorf("Error when unmarshalling the JSON file: %w", err)
	}

	return output, err
}

package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type ServerConfig struct {
	Server struct {
		Address string `yaml:"address"`
		Limits  struct {
			UploadRequests   int `yaml:"upload_requests"`
			DownloadRequests int `yaml:"download_requests"`
			ListRequests     int `yaml:"list_requests"`
		} `yaml:"limits"`
	} `yaml:"server"`
}

func LoadConfig(filePath string) (*ServerConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config ServerConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

package config

import (
	"gopkg.in/yaml.v3"
	"log"
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
	ServerDataDir string `yaml:"server_data_dir"`
}

func LoadConfig(filePath string) (*ServerConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("read config file failed. Err: %v", err)
		return nil, err
	}

	var config ServerConfig
	if err = yaml.Unmarshal(data, &config); err != nil {
		log.Printf("unmarshal config file failed. Err: %v", err)
		return nil, err
	}

	return &config, nil
}

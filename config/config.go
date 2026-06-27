package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

const configFilePath = "config.yaml"

var AppConfig Config

type Config struct {
	Server struct {
		Port  string `yaml:"port"`
		Debug bool   `yaml:"debug"`
	} `yaml:"server"`
	Database struct {
		Type string `yaml:"type"`
		DSN  string `yaml:"dsn"`
	} `yaml:"database"`
}

func LoadConfig() {
	file, err := os.Open(configFilePath)
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&AppConfig); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}
}
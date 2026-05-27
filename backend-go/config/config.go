package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type Config struct {
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
	} `json:"database"`
	GoServer struct {
		Port int `json:"port"`
	} `json:"go_server"`
	PythonServer struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"python_server"`
	JWT struct {
		Secret string `json:"secret"`
	} `json:"jwt"`
	Email struct {
		From     string `json:"from"`
		Password string `json:"password"`
		SMTPHost string `json:"smtp_host"`
		SMTPPort string `json:"smtp_port"`
	} `json:"email"`
	EmailVerificationEnabled bool `json:"email_verification_enabled"`
	Debug                    bool `json:"debug"`
}

var Global *Config

func Load() (*Config, error) {
	paths := []string{"settings.json", "../settings.json", "./settings.json"}
	var data []byte
	var err error
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("не удалось найти settings.json: %v", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	Global = &cfg

	if !cfg.Debug {
		log.SetOutput(io.Discard)
	} else {
		log.SetFlags(log.Ldate | log.Ltime)
	}

	return &cfg, nil
}

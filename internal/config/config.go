package config

import (
	"encoding/json"
	"os"
)

type DatabaseConfig struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Config struct {
	App struct {
		Mode      string   `json:"mode"`
		Key       []byte   `json:"key"`
		Languages []string `json:"languages"`
		Client    string   `json:"client"`
	} `json:"app"`
	Server struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"server"`
	Database DatabaseConfig `json:"database"`
	Mail     struct {
		FromName     string `json:"fromName"`
		FromEmail    string `json:"fromEmail"`
		SmtpServer   string `json:"smtpServer"`
		SmtpPort     string `json:"smtpPort"`
		SmtpSecurity string `json:"smtpSecurity"`
		SmtpUsername string `json:"smtpUsername"`
		SmtpPassword string `json:"smtpPassword"`
		AdminEmail   string `json:"adminEmail"`
	} `json:"mail"`
	Services struct {
	} `json:"services"`
}

var config *Config

func init() {
	config = loadConfig("../../config.json")
}

func loadConfig(file string) *Config {
	f, err := os.Open(file)
	if err != nil {
		panic("Wrong config path")
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		panic("Decoding config error")
	}

	return config
}

func Get() *Config {
	return config
}

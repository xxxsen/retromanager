package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type DBConfig struct {
	Host string
	Port uint32
	User string
	Pwd  string
	DB   string
}

type LogConfig struct {
	File      string `json:"file"`
	Level     string `json:"level"`
	FileSize  uint64 `json:"file_size"`
	FileCount uint64 `json:"file_count"`
	KeepDays  uint32 `json:"keep_days"`
	Console   bool   `json:"console"`
}

type ServerConfig struct {
	Address string
}

type Config struct {
	LogInfo    LogConfig    `json:"log_info"`
	DBInfo     DBConfig     `json:"db_info"`
	ServerInfo ServerConfig `json:"server_info"`
}

func Parse(f string) (*Config, error) {
	raw, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("read file:%w", err)
	}
	c := &Config{}
	if err := json.Unmarshal(raw, c); err != nil {
		return nil, fmt.Errorf("decode json:%w", err)
	}
	return c, nil
}

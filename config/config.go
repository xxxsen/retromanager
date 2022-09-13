package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/xxxsen/common/database"
)

type DBConfig struct {
	Host string `json:"host"`
	Port uint32 `json:"port"`
	User string `json:"user"`
	Pwd  string `json:"pwd"`
	DB   string `json:"db"`
}

type LogConfig struct {
	File      string `json:"file"`
	Level     string `json:"level"`
	FileSize  uint64 `json:"file_size"`
	FileCount uint64 `json:"file_count"`
	KeepDays  uint32 `json:"keep_days"`
	Console   bool   `json:"console"`
}

type S3Config struct {
	Endpoint  string `json:"endpoint"`
	SecretId  string `json:"secret_id"`
	SecretKey string `json:"secret_key"`
	UseSSL    bool   `json:"use_ssl"`
	Bucket    string `json:"bucket"`
}

type EsConfig struct {
	User     string   `json:"user"`
	Password string   `json:"password"`
	Timeout  int      `json:"timeout"`
	Host     []string `json:"host"`
}

type ServerConfig struct {
	Address string `json:"address"`
}

type IDGenConfig struct {
	WorkerID uint16 `json:"worker_id"`
}

type Config struct {
	LogInfo    LogConfig         `json:"log_info"`
	GameDBInfo database.DBConfig `json:"game_db_info"`
	FileDBInfo database.DBConfig `json:"file_db_info"`
	ServerInfo ServerConfig      `json:"server_info"`
	S3Info     S3Config          `json:"s3_info"`
	IDGenInfo  IDGenConfig       `json:"idgen_info"`
	EsInfo     EsConfig          `json:"es_info"`
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

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/xxxsen/common/database"
	"github.com/xxxsen/common/es"
	"github.com/xxxsen/common/logger"
)

type ServerConfig struct {
	Address string `json:"address"`
}

type IDGenConfig struct {
	WorkerID uint16 `json:"worker_id"`
}

type Config struct {
	LogInfo    logger.LogConfig  `json:"log_info"`
	GameDBInfo database.DBConfig `json:"game_db_info"`
	ServerInfo ServerConfig      `json:"server_info"`
	IDGenInfo  IDGenConfig       `json:"idgen_info"`
	EsInfo     es.EsConfig       `json:"es_info"`
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

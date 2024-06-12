package config

import (
	"encoding/json"
	"os"
)

type TGCreds struct {
	TgAPIKey  string `json:"tgAPIkey"`
	TgChannel string `json:"tgChannel"`
}

type Config struct {
	Telegram           TGCreds `json:"Telegram"`
	TrafikverketAPIKey string  `json:"TrafikverketAPIKey"`
	County             []int   `json:"County"`
}

var Loaded Config

func LoadConfig(filepath string) (Config, error) {
	var c Config
	data, err := os.ReadFile(filepath)
	if err != nil {
		return Config{}, err
	}
	err = json.Unmarshal(data, &c)
	if err != nil {
		return Config{}, err
	}
	Loaded = c

	return c, nil
}

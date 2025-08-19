package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

const envPrefix = "#env:"

type Secret string

func (s *Secret) UnmarshalJSON(data []byte) error {
	var key string
	if err := json.Unmarshal(data, &key); err != nil {
		return err
	}

	if !strings.HasPrefix(key, envPrefix) {
		*s = Secret(key)
		return nil
	}

	env := strings.TrimPrefix(key, envPrefix)
	value, exists := os.LookupEnv(env)
	if !exists {
		return fmt.Errorf("env var %s doesn't exist", env)
	}
	*s = Secret(value)

	return nil
}

func (s *Secret) Normalize() string {
	return string(*s)
}

type Config struct {
	ClientID      string `json:"client_id"`
	PersonalToken Secret `json:"personal_token"`
	AccountID     string `json:"account_id"`
	Addr          string `json:"addr"`
	ApiUrl        string `json:"api_url"`
}

func ParseConfig(path string) Config {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Cannot open config file: %v", err)
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		log.Fatalf("Cannot parse config file: %v", err)
	}

	return cfg
}

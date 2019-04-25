package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSL      string `json:"ssl"`
}

func (c PostgresConfig) Dialect() string {
	return "postgres"
}

func (c PostgresConfig) ConnectionInfo() string {
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", c.Host, c.Port, c.User, c.Name, c.SSL)
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", c.Host, c.Port, c.User, c.Password, c.Name, c.SSL)
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "sParhwk72",
		Name:     "lenslocked_dev",
		SSL:      "disable",
	}
}

type Config struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACKey  string         `json:"hmac_key"`
	Database PostgresConfig `json:"database"`
}

func (c Config) IsProd() bool {
	return strings.ToLower(c.Env) == "prod"
}

func DefaultConfig() Config {
	return Config{
		Port:     3000,
		Env:      "dev",
		Pepper:   "7SZ5t9epC5RFv&*",
		HMACKey:  "secret-key",
		Database: DefaultPostgresConfig(),
	}
}

func LoadConfig(configReq bool) Config {
	f, err := os.Open(".config")
	if err != nil {
		if configReq {
			panic(err)
		}
		log.Println("Using default config...")
		return DefaultConfig()
	}
	var c Config
	dec := json.NewDecoder(f)
	err = dec.Decode(&c)
	if err != nil {
		panic(err)
	}
	log.Println("Sucessfully loaded config from .config")
	return c
}

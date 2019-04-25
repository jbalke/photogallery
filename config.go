package main

import (
	"fmt"
	"strings"
)

// const (
// 	httpPort = ":3000"
// )

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
	SSL      string `json:"ssl"`
	Logging  bool   `json:"logging"`
}

func (c PostgresConfig) IsLogging() bool {
	return c.Logging
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
		Logging:  false,
	}
}

type Config struct {
	Port int
	Env  string
}

func (c Config) IsProd() bool {
	return strings.ToLower(c.Env) == "prod"
}

func DefaultConfig() Config {
	return Config{
		Port: 3000,
		Env:  "dev",
	}
}

// isProd := false

// const (
// 	userPwPepper      = "7SZ5t9epC5RFv&*"
// 	hmacSecretKey     = "secret-key"
// )

// db, err := gorm.Open("postgres", connectionInfo)
// 	if err != nil {
// 		return nil, err
// 	}
// 	db.LogMode(logging)

package config

import (
	"fmt"
	"sync"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Port        string `env:"PORT"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	DbUrl       string `env:"DATABASE_URL"`
	ReconnTime  int    `env:"RECONN_TIME" envDefault:"5"`
	ConnCheck   bool   `env:"CONN_CHECK" envDefault:"true"`
	ReconnTries int    `env:"RECONN_TRIES" envDefault:"5"`
	Redis       Redis
	CH          ClickHouseConfig
}
type Redis struct {
	Addr           string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	ExpTimeSeconds int    `env:"REDIS_EXP_TIME" envDefault:"60"`
	DBIndex        int    `env:"REDIS_DB_INDEX" envDefault:"0"`
}

type ClickHouseConfig struct {
	Addr string `env:"CH_ADDR" envDefault:"localhost:9000"`
	DB   string `env:"CH_DB"`
	User string `env:"CH_USER" envDefault:"default"`
	Pass string `env:"CH_PASS" envDefault:""`
}

var once sync.Once

var configInstance *Config

func GetConfig() *Config {
	if configInstance == nil {
		once.Do(func() {
			fmt.Println("Creating config instance now.")

			var cfg Config
			var redis Redis
			var ch ClickHouseConfig

			if err := env.Parse(&cfg); err != nil {
				log.Fatal(err)
			}
			if err := env.Parse(&redis); err != nil {
				log.Fatal(err)
			}
			if err := env.Parse(&ch); err != nil {
				log.Fatal(err)
			}
			cfg.Redis = redis
			cfg.CH = ch

			configInstance = &cfg
		})

	}

	return configInstance
}

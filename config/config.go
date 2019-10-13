package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Log struct {
	Path string
}

type Sqlite struct {
	Path string
}

type Server struct {
	Port string
}

type Config struct {
	Log
	Sqlite
	Server
}

func MakeConfig(v *viper.Viper) *Config {
	return &Config{
		Log: Log{
			Path: v.GetString("log.path"),
		},
		Sqlite: Sqlite{
			Path: v.GetString("sqlite.path"),
		},
		Server: Server{
			Port: v.GetString("server.port"),
		},
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("log.path", "stderr")

	v.SetDefault("sqlite.path", ":memory:")

	v.SetDefault("server.port", "9000")
}

// NewViper returns new configured *viper.Viper instance
func NewViper() (*viper.Viper, error) {
	v := viper.New()

	v.SetEnvPrefix("AE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	return v, nil
}

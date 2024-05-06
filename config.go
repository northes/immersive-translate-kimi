package main

import (
	"strings"

	"github.com/spf13/viper"
)

var ConfigPath = struct {
	Prompt string
	Port   string
	Key    string
}{
	Prompt: "PROMPT",
	Port:   "PORT",
	Key:    "KEY",
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	return nil
}

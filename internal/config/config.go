package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Telegram TelegramConfig `mapstructure:"telegram"`
	Database DatabaseConfig `mapstructure:"database"`
	Video    VideoConfig    `mapstructure:"video"`
	Server   ServerConfig   `mapstructure:"server"`
}

type TelegramConfig struct {
	Token     string `mapstructure:"token"`
	ChannelID int64  `mapstructure:"channel_id"`
	GroupLink string `mapstructure:"group_link"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type VideoConfig struct {
	MessageID int `mapstructure:"message_id"`
}

type ServerConfig struct {
	URL  string `mapstructure:"url"`
	Port string `mapstructure:"port"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	return &config
}

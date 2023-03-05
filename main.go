package main

import (
	"fmt"
	"github.com/spf13/viper"
)

var (
	config RedGifsConfig
)

const (
	ErrNoConfigFile       = "no app.env config file found"
	ErrNoHTTPPort         = "no HTTP port provided"
	ErrNoRedGifsClientID  = "no RedGifs client id provided"
	ErrNoRedGifsClientKey = "no RedGifs client secret provided"
)

type RedGifsConfig struct {
	HttpPort            string `mapstructure:"LISTEN_PORT"`
	RedGifsClientId     string `mapstructure:"REDGIFS_CLIENT_ID"`
	RedGifsClientSecret string `mapstructure:"REDGIFS_CLIENT_SECRET"`
}

func main() {
	tempConfig, err := loadConfig(".")
	if err != nil {
		panic(ErrNoConfigFile)
	}

	err = validateConfig(tempConfig)
	if err != nil {
		panic(err)
	}

	config = tempConfig
}

// loadConfig - Loads the config at a given path, returning the
// unmarshalled data as the return type.
func loadConfig(path string) (RedGifsConfig, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return RedGifsConfig{}, err
	}

	var rgConfig RedGifsConfig
	err = viper.Unmarshal(&rgConfig)
	if err != nil {
		return RedGifsConfig{}, err
	}

	return rgConfig, nil
}

// validateConfig - Checks if certain properties are present
func validateConfig(config RedGifsConfig) error {
	type errorEntry struct {
		name    string
		message string
	}

	var errors []errorEntry

	if len(config.HttpPort) == 0 {
		errors = append(errors, errorEntry{"HttpPort", ErrNoHTTPPort})
	}

	if len(config.RedGifsClientId) == 0 {
		errors = append(errors, errorEntry{"RedGifsClientId", ErrNoRedGifsClientID})
	}

	if len(config.RedGifsClientSecret) == 0 {
		errors = append(errors, errorEntry{"RedGifsClientSecret", ErrNoRedGifsClientKey})
	}

	if len(errors) == 0 {
		return nil
	}

	var errorBuilder string
	for _, entry := range errors {
		errorBuilder += fmt.Sprintf("%s: %s\n", entry.name, entry.message)
	}

	return fmt.Errorf("config validation failed:\n%s", errorBuilder)
}

package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
	"net/http"
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
	ListenPort          string `mapstructure:"LISTEN_PORT"`
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

	e := echo.New()
	e.GET("/redgifs/gif/:id", handleGifLookup)
	e.IPExtractor = echo.ExtractIPFromXFFHeader()
	e.Logger.Fatal(e.Start(config.ListenPort))
}

// handleGifLookup - Handles GET requests to send the stream URL to the client.
func handleGifLookup(c echo.Context) error {
	gifId := c.Param("id")
	return c.String(http.StatusOK, "You requested "+gifId)
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

// validateConfig - Checks if certain properties are present.
func validateConfig(config RedGifsConfig) error {
	type errorEntry struct {
		name    string
		message string
	}

	var errors []errorEntry

	if len(config.ListenPort) == 0 {
		errors = append(errors, errorEntry{"ListenPort", ErrNoHTTPPort})
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

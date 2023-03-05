package main

import "github.com/spf13/viper"

var (
	config RedGifsConfig
)

type RedGifsConfig struct {
	HttpPort            string `mapstructure:"HTTP_PORT"`
	RedGifsClientId     string `mapstructure:"REDGIFS_CLIENT_ID"`
	RedGifsClientSecret string `mapstructure:"REDGIFS_CLIENT_SECRET"`
}

func main() {
	tempConfig, err := loadConfig(".")
	if err != nil {
		panic("Could not load configuration file.")
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

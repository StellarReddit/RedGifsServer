package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/StellarReddit/RedGifsWrapper"
	"github.com/labstack/echo/v4"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	config     RedGifsConfig
	client     RedGifsWrapper.Client
	credential Credential
)

const (
	ErrNoConfigFile       = "no app.env config file found"
	ErrNoHTTPPort         = "no HTTP port provided"
	ErrNoRedGifsClientID  = "no RedGifs client id provided"
	ErrNoRedGifsClientKey = "no RedGifs client secret provided"
	ErrNoRedGifsTestId    = "no RedGifs test id provided"
	ServerUserAgent       = "app.stellarreddit.RedGifsServer (email: legal@azimuthcore.com)"
)

type Credential struct {
	accessTokenMutex sync.RWMutex
	accessToken      string
}

type RedGifsConfig struct {
	ListenPort          string `mapstructure:"LISTEN_PORT"`
	RedGifsClientId     string `mapstructure:"REDGIFS_CLIENT_ID"`
	RedGifsClientSecret string `mapstructure:"REDGIFS_CLIENT_SECRET"`
	RedGifsTestId       string `mapstructure:"REDGIFS_TEST_ID"`
}

type RedGifStreamUrlResponse struct {
	Url string `json:"streamUrl"`
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

	setupAccessTokenRefreshTask()
	setupRedGifsWrapperClient(tempConfig)

	e := echo.New()
	e.GET("/redgifs/gif/:id", handleGifLookup)
	e.IPExtractor = echo.ExtractIPFromXFFHeader()
	e.Logger.Fatal(e.Start(config.ListenPort))
}

// handleGifLookup - Handles GET requests to send the stream URL to the client.
func handleGifLookup(c echo.Context) error {
	gifId := c.Param("id")

	credential.accessTokenMutex.RLock()
	accessToken := credential.accessToken
	credential.accessTokenMutex.RUnlock()

	streamUrl, err := client.LookupStreamURL(c.RealIP(), "temp", gifId, accessToken)
	if errors.Is(err, RedGifsWrapper.ErrNotFound) {
		return c.String(http.StatusNotFound, "Could not find the stream url for the gif.")
	} else if err != nil {
		return c.String(http.StatusInternalServerError, "Something went wrong requesting the gif.")
	}

	var response RedGifStreamUrlResponse
	response.Url = streamUrl
	return c.JSON(http.StatusOK, response)
}

// setupAccessTokenRefreshTask - Run the refresh task on Saturdays at midnight
func setupAccessTokenRefreshTask() {
	c := cron.New()
	_, _ = c.AddFunc("@weekly", func() {
		attemptAccessTokenRefresh()
	})
	c.Start()
}

// setupRedGifsWrapperClient - Set up the RedGifs wrapper
func setupRedGifsWrapperClient(redGifsConfig RedGifsConfig) {
	redGifsWrapperConfig := RedGifsWrapper.Config{
		ClientID:     redGifsConfig.RedGifsClientId,
		ClientSecret: redGifsConfig.RedGifsClientSecret,
		UserAgent:    ServerUserAgent,
	}

	client = RedGifsWrapper.NewClient(redGifsWrapperConfig)
}

// attemptAccessTokenRefresh - Attempts to refresh the access token up to 5 times.
// Importantly, it validates tests the token is valid. Sometimes RedGifs issues
// broken tokens.
func attemptAccessTokenRefresh() {
	maxRetries := 5
	backoff := [5]time.Duration{5, 10, 30, 60, 120}

	for i := 0; i < maxRetries; i++ {
		accessToken, err := client.RequestNewAccessToken()

		if err != nil {
			time.Sleep(backoff[i] * time.Second)
			continue
		}

		// Wait for the token to become active
		time.Sleep(5 * time.Second)

		randomIp := generateRandomIPv4Address()
		_, err = client.LookupStreamURL(randomIp, ServerUserAgent, config.RedGifsTestId, accessToken)

		if err != nil {
			time.Sleep(backoff[i] * time.Second)
			continue
		}

		credential.accessTokenMutex.Lock()
		credential.accessToken = accessToken
		credential.accessTokenMutex.Unlock()
		break
	}
}

// generateRandomIPv4Address - generate a random IPv4 address for testing
// access tokens.
func generateRandomIPv4Address() string {
	buf := make([]byte, 4)
	ip := rand.Uint32()
	binary.LittleEndian.PutUint32(buf, ip)
	return fmt.Sprintf("%s", net.IP(buf))
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

	var entries []errorEntry

	if len(config.ListenPort) == 0 {
		entries = append(entries, errorEntry{"ListenPort", ErrNoHTTPPort})
	}

	if len(config.RedGifsClientId) == 0 {
		entries = append(entries, errorEntry{"RedGifsClientId", ErrNoRedGifsClientID})
	}

	if len(config.RedGifsClientSecret) == 0 {
		entries = append(entries, errorEntry{"RedGifsClientSecret", ErrNoRedGifsClientKey})
	}

	if len(config.RedGifsTestId) == 0 {
		entries = append(entries, errorEntry{"RedGifsTestId", ErrNoRedGifsTestId})
	}

	if len(entries) == 0 {
		return nil
	}

	var errorBuilder string
	for _, entry := range entries {
		errorBuilder += fmt.Sprintf("%s: %s\n", entry.name, entry.message)
	}

	return fmt.Errorf("config validation failed:\n%s", errorBuilder)
}

package main

import (
	"blockmesh/constant"
	"blockmesh/request"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/mattn/go-colorable"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"math/rand"
	"regexp"
	"sync"
	"time"
)

var lock struct {
	sync.Mutex // <-- this mutex protects
}

var logger *zap.Logger

func main() {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	))

	viper.SetConfigFile("./conf.toml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	proxies := viper.GetStringSlice("proxies.data")

	var accounts []request.LoginRequest
	err = viper.UnmarshalKey("data.auth", &accounts)
	if err != nil {
		logger.Error("Error unmarshalling config: %v\n", zap.Error(err))
		return
	}

	for i, acc := range accounts {
		go ping(proxies[i%len(proxies)], acc)
	}

	select {}

}

func ping(proxyURL string, authInfo request.LoginRequest) {
	rand.Seed(time.Now().UnixNano())
	client := resty.New().SetProxy(proxyURL)
	lastLogin := time.Now()
	var loginResponse request.LoginResponse

	_, err := client.R().
		SetBody(authInfo).
		SetResult(&loginResponse).
		Post(constant.LoginURL)
	if err != nil {
		logger.Error("Login error", zap.String("email", authInfo.Email))
		time.Sleep(5 * time.Minute)
		go ping(proxyURL, authInfo)
		return
	}

	var publicIp *request.GetIPResponse

	_, err = client.R().
		SetResult(&publicIp).
		Get("https://api.ipify.org?format=json")
	if err != nil {
		panic("Can't get public ip")
	}

	var ipInformation request.IpInformation
	_, err = client.R().
		SetResult(&ipInformation).
		Get(fmt.Sprintf("https://ipinfo.io/%v/json", publicIp.IP))
	if err != nil {
		panic("Can't get ip information")
	}

	for {
		if time.Now().Sub(lastLogin) > time.Hour {
			_, err := client.R().
				SetBody(authInfo).
				SetResult(&loginResponse).
				Post(constant.LoginURL)
			if err != nil {
				logger.Error("Login error", zap.String("email", authInfo.Email))
				time.Sleep(5 * time.Minute)
				go ping(proxyURL, authInfo)
				return
			}
			lastLogin = time.Now()
		}
		payload := map[string]interface{}{
			"email":     authInfo.Email,
			"api_token": loginResponse.APIToken,
		}

		res, err := client.R().
			SetBody(payload).
			Post(constant.TaskURL)

		// Check for errors
		if err != nil {
			logger.Error("Error getting task request: ", zap.Error(err))
		}
		logger.Info("Getting task request successfully", zap.String("email", authInfo.Email), zap.Any("res", res))

		// Generate a random float between 12 and 14
		min := 150.0
		max := 180.0
		downloadSpeed := min + (max-min)*rand.Float64()

		min = 42.0
		max = 46.0

		// Generate a random float between min and max
		latency := min + (max-min)*rand.Float64()

		min = 8.0
		max = 11.0

		// Generate a random float between min and max
		uploadSpeed := min + (max-min)*rand.Float64()

		re := regexp.MustCompile(`\d+`)

		// Find the first match of digits in the input string
		match := re.FindString(ipInformation.Org)

		// Define the request payload
		payload = map[string]interface{}{
			"email":          authInfo.Email,
			"api_token":      loginResponse.APIToken,
			"download_speed": downloadSpeed,
			"upload_speed":   uploadSpeed,
			"latency":        latency,
			"city":           ipInformation.City,
			"country":        ipInformation.Country,
			"ip":             ipInformation.IP,
			"asn":            match,
			"colo":           "NYC",
		}

		res, err = client.R().
			SetBody(payload).
			Post(constant.BandwidthURL)

		// Check for errors
		if err != nil {
			logger.Error("Error submitting bandwidth request: ", zap.Error(err))
		}
		logger.Info("Submitting bandwidth request successfully", zap.String("email", authInfo.Email), zap.Any("res", res))

		res, err = client.R().
			SetQueryParams(map[string]string{
				"email":     authInfo.Email,
				"api_token": loginResponse.APIToken,
				"ip":        publicIp.IP,
			}).
			Post(constant.UptimeURL)

		// Check for errors
		if err != nil {
			logger.Error("Error making uptime request: ", zap.Error(err))
		}
		logger.Info("Making uptime request successfully", zap.String("email", authInfo.Email), zap.Any("res", res))
		time.Sleep(30 * time.Second)
	}
}

package main

import (
	"blockmesh/request"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

var logger *zap.Logger

func main() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, _ = config.Build()
	// Create a launcher with options
	path, _ := launcher.LookPath()

	extensionPath := "./extensions/blockmesh"

	type Proxy struct {
		Res []string `json:"res"`
	}

	var proxies Proxy

	client := resty.New()
	_, err := client.R().SetResult(&proxies).Get("http://localhost:3000")
	if err != nil {
		panic("Get proxy failed")
	}

	viper.SetConfigFile("./conf.toml")
	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	var accounts []request.LoginRequest
	err = viper.UnmarshalKey("data.auth", &accounts)
	if err != nil {
		logger.Error("Error unmarshalling config: %v\n", zap.Error(err))
		return
	}

	for i, acc := range accounts {
		l := launcher.New().Bin(path).
			Headless(false).
			NoSandbox(true).
			Set("disable-web-security").
			Set("disable-site-isolation-trials").
			Set("load-extension", extensionPath)
		l = l.Proxy(proxies.Res[i])

		// Launch a browser instance
		browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()

		// Ensure the browser is closed at the end
		defer browser.MustClose()

		page := browser.MustPage("https://app.blockmesh.xyz/ext/login")

		// Enable network events
		_ = proto.NetworkEnable{}.Call(page)

		// Log all responses
		go page.EachEvent(func(e *proto.NetworkResponseReceived) {
			fmt.Printf("Response Status: %d\n", e.Response.Status)
			// Fetch and log the response body
			body, err := proto.NetworkGetResponseBody{RequestID: e.RequestID}.Call(page)
			if err != nil {
				log.Printf("Failed to get response body for %s: %v\n", e.Response.URL, err)
			} else {
				fmt.Printf("Response Body: %s\n", body.Body)
			}
		})

		// Find the iframe element by its CSS selector (e.g., using the iframe's name or ID)
		page.MustElement("input").MustWaitLoad()

		inputs := page.MustElements("input")
		inputs[0].MustFocus()
		inputs[0].MustInput(acc.Email)

		inputs[1].MustFocus()
		inputs[1].MustInput(acc.Password)

		btn := page.MustElement("button")
		btn.MustClick()

		log.Println("Chrome with extension loaded successfully!")
	}

	select {}
}

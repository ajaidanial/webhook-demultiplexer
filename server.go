package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Configuration defines the single config dict from the config.json file.
type Configuration struct {
	Host    string   `json:"host"`
	Targets []string `json:"targets"`
}

/**
This file opens the `config.json` file and reads the contents and returns the data.
This is used by other functions to get the configurations.
*/
func getConfigurations() []Configuration {
	configFile, error := os.Open("config.json")

	// error while opening the file
	if error != nil {
		fmt.Println("Error while opening config file.", error)
		os.Exit(3)
	}

	// read the file and get contents
	defer configFile.Close()
	byteValue, _ := ioutil.ReadAll(configFile)
	var configurations []Configuration
	json.Unmarshal(byteValue, &configurations)

	return configurations
}

func main() {
	// server instance
	server := echo.New()

	// middlewares & config
	server.Pre(middleware.AddTrailingSlash())

	// url mapping
	server.Any("/webhook/", webhookHandler)
	server.GET("/ping/", pingHandler)

	// running the server
	server.Logger.Fatal(server.Start(":8080"))
}

/**
Handler for the `/ping` endpoint. This is used as a status check for the application.
Like, if in case this is needed for anything.
*/
func pingHandler(context echo.Context) error {
	return context.String(http.StatusOK, "pong")
}

/**
Main handler for all the webhooks. This is called using any method, from any where.
Based on the config.json file this endpoint handles and demultiplexes the requests.
*/
func webhookHandler(context echo.Context) error {
	request := context.Request()

	// inbound data
	triggeredHost := request.Host
	triggeredMethod := request.Method

	// check given config based on inbound data
	for _, config := range getConfigurations() {

		if config.Host == triggeredHost {
			targetsToHit := config.Targets
			for _, target := range targetsToHit {
				fmt.Println(target, triggeredMethod)
			}
		}
	}

	return context.String(http.StatusOK, "webhookHandler - working")
}

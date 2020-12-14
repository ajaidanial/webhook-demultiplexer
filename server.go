package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Configuration defines the single config dict from the config.json file.
type Configuration struct {
	Host    string   `json:"host"`
	Targets []string `json:"targets"`
}

// MessageResponse defines the output schema for all the endpoints.
type MessageResponse struct {
	Message string `json:"message"`
}

/**
This file opens the `config.json` file and reads the contents and returns the data.
This is used by other functions to get the configurations.
*/
func getConfigurations() []Configuration {
	configFileName := "config.json"
	configFile, error := os.Open(configFileName)

	// error while opening the file
	if error != nil {
		fmt.Println("Error while opening configuration file.", error)
		fmt.Printf("Make sure the `%s` file is present in the project root.\n", configFileName)
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
	return context.JSON(http.StatusOK, MessageResponse{"pong"})
}

/**
Main handler for all the webhooks. This is called using any method, from any where.
Based on the config.json file this endpoint handles and demultiplexes the requests.
*/
func webhookHandler(context echo.Context) error {
	request := context.Request()

	// getting the reponse body & preparing for request
	inboundData := echo.Map{}
	if error := context.Bind(&inboundData); error != nil {
		return error
	}
	byteArrayData, _ := json.Marshal(inboundData)

	// check given config based on inbound data
	for _, config := range getConfigurations() {

		if config.Host == request.Host {
			for _, target := range config.Targets {

				// prepare the request to be sent
				outboundRequest, _ := http.NewRequest(
					request.Method,
					target,
					bytes.NewBuffer(byteArrayData),
				)
				for headerKey, headerValues := range request.Header {
					outboundRequest.Header.Set(headerKey, strings.Join(headerValues, ", "))
				}
				fmt.Println("Outbound Request: \n", outboundRequest)

				// send the request and get the response
				client := &http.Client{}
				response, error := client.Do(outboundRequest)
				if error != nil {
					panic(error)
				}
				defer response.Body.Close()

				fmt.Println("Response: \n", response)
				responseBody, _ := ioutil.ReadAll(response.Body)
				fmt.Println("Response Body: \n", string(responseBody))
				fmt.Println("-------------- XXXX --------------")
			}
		}
	}

	return context.JSON(http.StatusOK, MessageResponse{"done"})
}

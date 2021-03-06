package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

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
	configFile, err := os.Open(configFileName)

	// error while opening the file
	if err != nil {
		fmt.Println("Error while opening configuration file.", err)
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
	// check if everything is fine by getting the configurations
	getConfigurations()

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

	// getting the response body
	inboundData := echo.Map{}
	err := context.Bind(&inboundData)
	checkAndPanic(err)

	// check given config based on inbound data | de-multiplexer operation
	for _, config := range getConfigurations() {
		if config.Host == request.Host || config.Host == "*" {
			for _, target := range config.Targets {
				// forward the request
				forwardRequest(target, inboundData, request)
			}
		}
	}

	return context.JSON(http.StatusOK, MessageResponse{"done"})
}

/**
This is the main function used to forward the inbound values to specified targets in the
the config.json file. This is called from the handler for the `/webhook/` endpoint.
*/
func forwardRequest(target string, inboundData echo.Map, inboundRequest *http.Request) {

	var processedOutputData io.Reader

	// if body is present, process the data, else let it be nil
	if inboundData != nil && len(inboundData) > 0 {
		byteArrayData, _ := json.Marshal(inboundData)
		processedOutputData = bytes.NewBuffer(byteArrayData)
	}

	// prepare the request to be sent
	outboundRequest, _ := http.NewRequest(
		inboundRequest.Method,
		target,
		processedOutputData,
	)
	for headerKey, headerValues := range inboundRequest.Header {
		outboundRequest.Header.Set(headerKey, strings.Join(headerValues, ", "))
	}

	client := &http.Client{}
	response, err := client.Do(outboundRequest)
	checkAndPanic(err)
	defer response.Body.Close()
	responseBody, _ := ioutil.ReadAll(response.Body)

	// make the string log
	log := fmt.Sprintf(
		"\nFrom: %s \nTarget: %s \nResponse Code: %d \nResponse Body: %s \nTime: %s \n",
		inboundRequest.Host,
		target,
		response.StatusCode,
		string(responseBody),
		time.Now().String(),
	)

	// console log
	fmt.Println(log)
	// file log
	file, _ := os.OpenFile("events.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	defer file.Close()
	file.WriteString(log)
}

/**
This function checks if there are errors present. If so, calls the `panic()` function.
This is defined as a function to make things DRY.
*/
func checkAndPanic(e error) {
	if e != nil {
		panic(e)
	}
}

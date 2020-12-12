package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	// server instance
	server := echo.New()

	// url mapping
	server.GET("/ping", pingHandler)

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

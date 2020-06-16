package main

import (
	"log"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	basicAuthUserName := os.Getenv("USER_NAME")
	if basicAuthUserName == "" {
		log.Fatal("missing USER_NAME environment variables")
	}
	basicAuthPassword := os.Getenv("PASSWORD")
	if basicAuthPassword == "" {
		log.Fatal("missing PASSWORD environment variables")
	}
	e := echo.New()
	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		return (username == basicAuthUserName && password == basicAuthPassword), nil
	}))
	e.POST("/cce", dispatchCCEHandler)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("server online at ", port)
	log.Fatal(e.Start(":" + port))
}

package main

import (
	"log"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	userName := os.Getenv("USER_NAME")
	if userName == "" {
		log.Fatal("missing username on environment variables")
	}
	password := os.Getenv("PASSWORD")
	if password == "" {
		log.Fatal("missing password on environment variables")
	}
	e := echo.New()
	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username != userName || password != password {
			return false, nil
		}
		return true, nil
	}))
	e.POST("/cce", dispatchCCEHandler)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("server online at ", port)
	log.Fatal(e.Start(":" + port))
}

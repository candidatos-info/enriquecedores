package main

import (
	"log"
	"os"

	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	e.POST("/cce", dispatchCCEHandler)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("server online at ", port)
	log.Fatal(e.Start(":" + port))
}

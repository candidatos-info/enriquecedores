package main

import (
	"log"
	"os"

	"github.com/candidatos-info/enriquecedores/cce"
	"github.com/labstack/echo"
)

func main() {
	e := echo.New()
	cceHandler := cce.NewHandler()
	e.POST("/cce", cceHandler.Post)
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("server online at ", port)
	log.Fatal(e.Start(":" + port))
}

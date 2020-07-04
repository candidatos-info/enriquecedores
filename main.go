package main

import (
	"fmt"
	"log"
	"os"

	"github.com/candidatos-info/enriquecedores/cce"
	"github.com/candidatos-info/enriquecedores/filestorage"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	baseDir := os.Getenv("CCE_BASE_DIR")
	if baseDir == "" {
		log.Fatal("missing CCE_BASE_DIR environment variable")
	}
	basicAuthUserName := os.Getenv("USER_NAME")
	if basicAuthUserName == "" {
		log.Fatal("missing USER_NAME environment variable")
	}
	basicAuthPassword := os.Getenv("PASSWORD")
	if basicAuthPassword == "" {
		log.Fatal("missing PASSWORD environment variable")
	}
	e := echo.New()
	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		return (username == basicAuthUserName && password == basicAuthPassword), nil
	}))
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing PORT environment variable")
	}
	fileStorage, err := filestorage.New()
	if err != nil {
		log.Fatal(fmt.Sprintf("falha ao criar cliente do Google Cloud Storage, erro %q", err))
	}
	cceHandler := cce.New(baseDir, fileStorage)
	e.POST("/cce", cceHandler.Post)
	e.GET("/cce", cceHandler.Get)
	log.Println("server online at ", port)
	log.Fatal(e.Start(":" + port))
}

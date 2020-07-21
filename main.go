package main

import (
	"log"
	"os"

	"github.com/candidatos-info/enriquecedores/candidatures"
	"github.com/candidatos-info/enriquecedores/filestorage"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	baseDir := os.Getenv("CANDIDATURES_BASE_DIR")
	if baseDir == "" {
		log.Fatal("missing CANDIDATURES_BASE_DIR environment variable")
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
	gcsClient, err := filestorage.NewGCSClient()
	if err != nil {
		log.Fatalf("falha ao criar cliente do Google Cloud Storage, erro %q", err)
	}
	cceHandler := candidatures.New(baseDir, gcsClient)
	e.POST("/candidatures", cceHandler.Post)
	e.GET("/candidatures", cceHandler.Get)
	log.Println("server online at ", port)
	log.Fatal(e.Start(":" + port))
}

package main

import (
	"log"
	"os"

	"github.com/candidatos-info/enriquecedores/cce"
	"github.com/labstack/echo"
)

func main() {
	filesURL := os.Getenv("FILES_URL")
	if filesURL == "" {
		log.Fatal("missing files URL on environment variables")
	}
	baseDir := os.Getenv("CCE_BASE_DIR")
	if baseDir == "" {
		log.Fatal("missing cce base dir on environment variables")
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("missing server port on environment variables")
	}
	cceHandler := cce.New(filesURL, baseDir)
	e := echo.New()
	e.POST("/cce", cceHandler.Post)
	e.GET("/cce", cceHandler.Get)
	log.Println("server online at ", port)
	log.Fatal(e.Start(":" + port))
}

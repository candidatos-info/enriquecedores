package cce

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

var (
	// hostURL can be changed by TestPost function test for tests purposes
	hostURL = "http://agencia.tse.jus.br/estatistica/sead/odsele/consulta_cand/consulta_cand_%d.zip"
)

const (
	outputPath = "sheets.zip"
)

// Handler is struct for the methods
type Handler struct {
}

// NewHandler does
func NewHandler() *Handler {
	return &Handler{}
}

type dispatchRequest struct {
	Year int64 `json:"year"`
}

// Post should be called to dispatch the process
func (h *Handler) Post(c echo.Context) error {
	in := dispatchRequest{}
	fmt.Println(c)
	err := c.Bind(&in)
	if err != nil {
		log.Println(fmt.Sprintf("failed to bind request input: %q", err))
		payload := make(map[string]string)
		payload["message"] = "Invalid request body"
		return c.JSON(http.StatusUnprocessableEntity, payload)
	}
	// TODO make tests easier se essa url for fixa os testes ficam dificultados
	downloadURL := fmt.Sprintf(hostURL, in.Year)
	f, err := os.Create(outputPath)
	if err != nil {
		log.Println(fmt.Sprintf("failed to create sheets zip file, got %q", err))
	}
	err = donwloadFile(downloadURL, f)
	if err != nil {
		log.Panicln(fmt.Sprintf("failed to download sheets, got %q", err))
	}
	return c.JSON(http.StatusOK, "ok")
}

// download a file and writes on the given writer
func donwloadFile(url string, w io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading file:%q", err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("error copying response content:%q", err)
	}
	return nil
}

package cce

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo"
)

// Handler is struct for the methods
type Handler struct {
}

// NewHandler does
func NewHandler() *Handler {
	return &Handler{}
}

// Post should be called to dispatch the process
func (h *Handler) Post(c echo.Context) error {
	return c.JSON(http.StatusOK, "nil")
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

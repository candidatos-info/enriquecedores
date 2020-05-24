package cce

import (
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

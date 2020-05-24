package main

import (
	"net/http"

	"github.com/labstack/echo"
)

func dispatchCCEHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, "hello-world")
}

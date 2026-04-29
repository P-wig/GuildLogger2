// backend/app/routes/root.go
package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterRoot(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"service": "backend",
			"status":  "ok",
		})
	})
}

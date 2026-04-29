// backend/app/routes/health.go
package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterHealth(e *echo.Echo) {
	e.GET("/api/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]bool{"ok": true})
	})
}

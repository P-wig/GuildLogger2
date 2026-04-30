// app/app.go
package app

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/P-wig/GuildLogger2/backend/app/config"
	"github.com/P-wig/GuildLogger2/backend/app/db"
	"github.com/P-wig/GuildLogger2/backend/app/routes"
)

// CreateApp builds and configures the Echo web server instance for the backend.
//
// Responsibilities in this file:
// 1. Load runtime configuration (Mongo URI/DB, CORS origins).
// 2. Create and configure the Echo server.
// 3. Register global middleware:
//   - Panic recovery middleware.
//   - Structured request logging middleware.
//   - CORS policy for frontend access.
//
// 4. Initialize MongoDB connection and return startup error if DB is unavailable.
// 5. Register HTTP routes (currently root and health).
// 6. Return a cleanup function that gracefully closes the Mongo client.
//
// CreateApp returns:
// - *echo.Echo: the configured server instance ready to start.
// - func() error: shutdown cleanup callback for DB resource disposal.
// - error: non-nil when startup wiring (for example DB init) fails.
func CreateApp() (*echo.Echo, func() error, error) {
	cfg := config.Load()

	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogMethod:   true,
		LogRemoteIP: true,
		LogLatency:  true,
		LogError:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				c.Logger().Infof("method=%s uri=%s status=%d latency=%s ip=%s",
					v.Method, v.URI, v.Status, v.Latency, v.RemoteIP)
			} else {
				c.Logger().Errorf("method=%s uri=%s status=%d latency=%s ip=%s err=%v",
					v.Method, v.URI, v.Status, v.Latency, v.RemoteIP, v.Error)
			}
			return nil
		},
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.CORSOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
	}))

	mongoClient, _, err := db.InitMongo(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		return nil, nil, err
	}

	// Route registration mirrors Flask blueprints.
	routes.RegisterRoot(e)
	routes.RegisterHealth(e)

	cleanup := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return db.CloseMongo(ctx, mongoClient)
	}

	return e, cleanup, nil
}

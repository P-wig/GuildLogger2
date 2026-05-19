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
	"github.com/P-wig/GuildLogger2/backend/app/discord"
	appmiddleware "github.com/P-wig/GuildLogger2/backend/app/middleware"
	"github.com/P-wig/GuildLogger2/backend/app/repositories"
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
// Route protection:
//
//   - Non-protected routes are accessible by anyone without a valid session.
//     Use these for public endpoints: login, OAuth callback, health checks.
//     Example: /api/auth/discord/login, /api/health
//
//   - Protected routes require a valid JWT in the Authorization header.
//     Use these for any endpoint that needs to know who the user is,
//     such as fetching user profile, listing guilds, logging events.
//     Example: /api/me, /api/guilds, /api/events
//
//     To add a new protected route, register it on the `protected` group:
//     protected.GET("/me", handler)
//     To add a new non-protected route, register it directly on `e` or an unguarded group:
//     e.GET("/health", handler)
//
// CreateApp returns:
// - *echo.Echo: the configured server instance ready to start.
// - func() error: shutdown cleanup callback for DB resource disposal.
// - error: non-nil when startup wiring (for example DB init) fails.
func CreateApp() (*echo.Echo, func() error, error) {
	cfg := config.Load()

	if err := config.ValidateOAuthConfig(cfg); err != nil {
		return nil, nil, err
	}

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

	mongoClient, database, err := db.InitMongo(cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		return nil, nil, err
	}

	userRepo := repositories.NewMongoUserRepository(database)
	memberRepo := repositories.NewMongoMemberRepository(database)
	eventsRepo := repositories.NewMongoEventRepository(database)
	guildRepo := repositories.NewMongoGuildRepository(database)

	// Ensure required user/member/event indexes exist at startup.
	idxCtx, idxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer idxCancel()

	// EnsureIndexes is called at startup so required DB constraints and query indexes
	// are present before handling traffic. This prevents duplicate identity records
	// (user/member uniqueness) and keeps common lookups fast (for example by guildId).
	// Creating indexes is idempotent, so calling this on every startup is safe.
	if err := userRepo.EnsureIndexes(idxCtx); err != nil {
		return nil, nil, err
	}
	if err := memberRepo.EnsureIndexes(idxCtx); err != nil {
		return nil, nil, err
	}
	if err := eventsRepo.EnsureIndexes(idxCtx); err != nil {
		return nil, nil, err
	}
	if err := guildRepo.EnsureIndexes(idxCtx); err != nil {
		return nil, nil, err
	}

	oauthClient := discord.NewOAuthClient(
		cfg.DiscordClientID,
		cfg.DiscordClientSecret,
		cfg.DiscordAuthBaseURL,
		cfg.DiscordAPIBaseURL,
		cfg.DiscordOAuthScopes,
	)

	// jwtMiddleware guards any route group that requires an authenticated session.
	// Routes registered on `protected` will reject requests without a valid JWT.
	jwtMiddleware := appmiddleware.JWTMiddleware(cfg)
	protected := e.Group("/api", jwtMiddleware)

	// Non-protected routes: no JWT required, accessible to anyone.
	routes.RegisterRoot(e)
	routes.RegisterHealth(e)
	routes.RegisterAuth(e, userRepo, oauthClient, cfg.SecretKey)

	// Protected routes: JWT middleware is enforced by the `protected` group.
	// Add new authenticated endpoints here as new route files are created.
	routes.RegisterAuthProtected(protected, userRepo)
	routes.RegisterGuildsProtected(protected, guildRepo, memberRepo, eventsRepo)
	routes.RegisterEventsProtected(protected, eventsRepo)
	routes.RegisterMembersProtected(protected, memberRepo)

	cleanup := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return db.CloseMongo(ctx, mongoClient)
	}

	return e, cleanup, nil
}

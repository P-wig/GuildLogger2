// app/app.go
package app

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	gommonlog "github.com/labstack/gommon/log"

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
	e.Logger.SetLevel(gommonlog.INFO)

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
	eventReportRepo := repositories.NewMongoEventReportRepository(database)
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
	if err := eventReportRepo.EnsureIndexes(idxCtx); err != nil {
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

	botClient := discord.NewBotClient(cfg.DiscordBotToken, cfg.DiscordAPIBaseURL)

	// Register bot slash commands with Discord on every startup.
	// PUT /applications/{id}/commands is a bulk overwrite — idempotent and safe to repeat.
	// A missing bot token (for example in environments without a bot configured) is
	// silently skipped. Any other failure is logged as a warning; it does not prevent
	// the server from starting.
	{
		cmdCtx, cmdCancel := context.WithTimeout(context.Background(), 15*time.Second)
		if err := botClient.RegisterGlobalCommands(cmdCtx, cfg.DiscordClientID); err != nil {
			e.Logger.Warnf("failed to register Discord slash commands: %v", err)
		} else if cfg.DiscordBotToken != "" {
			e.Logger.Infof("Discord slash commands registered (%d commands)", len(discord.BotCommands))
		}
		cmdCancel()
	}

	// eventService owns the event lifecycle shared by the Discord bot and the REST API.
	eventService := discord.NewEventService(eventsRepo, eventReportRepo, guildRepo, botClient)

	// Register Discord interaction webhook.
	// POST /api/interactions handles slash command responses, modal submits, and button clicks.
	// Authenticated via Ed25519 signature (not JWT) — must be registered before the JWT group.
	interactionHandler := discord.NewInteractionHandler(guildRepo, memberRepo, eventsRepo, eventReportRepo, botClient, eventService, cfg.DiscordPublicKey, cfg.SecretKey, cfg.AppURL)
	routes.RegisterInteractions(e, interactionHandler)

	// jwtMiddleware guards any route group that requires an authenticated session.
	// Routes registered on `protected` will reject requests without a valid JWT.
	jwtMiddleware := appmiddleware.JWTMiddleware(cfg)
	protected := e.Group("/api", jwtMiddleware)

	// Non-protected routes: no JWT required, accessible to anyone.
	routes.RegisterRoot(e)
	routes.RegisterHealth(e)
	routes.RegisterAuth(e, userRepo, oauthClient, cfg.SecretKey)
	routes.RegisterEventLogRoutes(e, eventsRepo, eventReportRepo, memberRepo, guildRepo, botClient, cfg.SecretKey)

	// Protected routes: JWT middleware is enforced by the `protected` group.
	// Add new authenticated endpoints here as new route files are created.
	routes.RegisterAuthProtected(protected, userRepo)
	routes.RegisterGuildsProtected(protected, guildRepo, memberRepo, eventsRepo, eventReportRepo, userRepo, oauthClient, botClient)
	routes.RegisterEventsProtected(protected, eventsRepo, eventReportRepo, guildRepo, memberRepo, botClient, eventService)
	routes.RegisterMembersProtected(protected, memberRepo)
	routes.RegisterNotificationsProtected(protected, guildRepo, memberRepo, botClient, eventsRepo)

	// Start the hourly reminder scheduler. It aligns its first tick to the top of
	// the next clock hour so reminders always fire on the hour boundary regardless
	// of when the server was started. schedulerCancel is called in cleanup to stop
	// the goroutine cleanly on shutdown.
	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	routes.StartReminderScheduler(schedulerCtx, eventsRepo, botClient, e.Logger)

	cleanup := func() error {
		schedulerCancel()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return db.CloseMongo(ctx, mongoClient)
	}

	return e, cleanup, nil
}

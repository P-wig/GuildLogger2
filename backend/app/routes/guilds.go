package routes

import (
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/discord"
	"github.com/P-wig/GuildLogger2/backend/app/repositories"
	"github.com/P-wig/GuildLogger2/backend/app/session"
	"github.com/labstack/echo/v4"
)

type connectGuildPayload struct {
	GuildID string `json:"guildId"`
	Name    string `json:"name"`
}

type upsertRolePayload struct {
	DiscordRoleID string                     `json:"discordRoleId"`
	Position      int                        `json:"position"`
	Type          repositories.GuildRoleType `json:"type"`
	Managed       bool                       `json:"managed"`
	IsDefault     bool                       `json:"isDefault"`
}

// RegisterGuildsProtected registers all guild routes on the JWT-guarded group.
func RegisterGuildsProtected(
	g *echo.Group,
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	eventRepo repositories.EventRepository,
	eventReportRepo repositories.EventReportRepository,
	userRepo repositories.UserRepository,
	oauthClient *discord.OAuthClient,
	botClient *discord.BotClient,
) {
	g.GET("/guilds", listGuildsHandler(guildRepo))
	g.GET("/guilds/discord", listDiscordGuildsHandler(userRepo, oauthClient))
	g.POST("/guilds/connect", connectGuildHandler(guildRepo, userRepo, oauthClient))
	g.POST("/guilds/:guildId/bot/install", installBotHandler(guildRepo))
	g.POST("/guilds/:guildId/bot/verify", verifyBotInstallHandler(guildRepo, botClient))
	g.GET("/guilds/:guildId/members/sync-status", memberSyncStatusHandler(guildRepo, memberRepo))
	g.POST("/guilds/:guildId/members/sync", syncMembersHandler(guildRepo, memberRepo, botClient))
	g.GET("/guilds/:guildId/dashboard", guildDashboardHandler(guildRepo, memberRepo, eventRepo, eventReportRepo))
	g.GET("/guilds/:guildId/bot/invite-url", botInviteURLHandler(guildRepo, oauthClient))
	g.PUT("/guilds/:guildId/roles", upsertRoleHandler(guildRepo))
}

func listGuildsHandler(guildRepo repositories.GuildRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guilds, err := guildRepo.FindByOwnerDiscordID(c.Request().Context(), claims.DiscordID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "guilds": guilds})
	}
}

func connectGuildHandler(guildRepo repositories.GuildRepository, userRepo repositories.UserRepository, oauthClient *discord.OAuthClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		var in connectGuildPayload
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}

		in.GuildID = strings.TrimSpace(in.GuildID)
		in.Name = strings.TrimSpace(in.Name)
		if in.GuildID == "" || in.Name == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId and name are required"})
		}

		ctx := c.Request().Context()

		user, err := userRepo.FindByDiscordID(ctx, claims.DiscordID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if user == nil || user.AccessToken == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "no Discord access token"})
		}

		discordGuilds, err := oauthClient.GetUserGuilds(ctx, user.AccessToken)
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]interface{}{"ok": false, "error": "failed to verify guild membership"})
		}

		isMember := false
		for _, dg := range discordGuilds {
			if dg.ID == in.GuildID {
				isMember = true
				break
			}
		}
		if !isMember {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "you are not a member of this Discord guild"})
		}

		guild := &repositories.Guild{
			GuildID:        in.GuildID,
			Name:           in.Name,
			OwnerDiscordID: claims.DiscordID,
		}

		if err := guildRepo.Create(ctx, guild); err != nil {
			if errors.Is(err, repositories.ErrGuildAlreadyExists) {
				return c.JSON(http.StatusConflict, map[string]interface{}{"ok": false, "error": "guild is already connected"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to connect guild"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "guild": guild})
	}
}

func installBotHandler(guildRepo repositories.GuildRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		guild, err := guildRepo.FindByGuildID(c.Request().Context(), guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}
		if guild.OwnerDiscordID != claims.DiscordID {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "you do not own this guild"})
		}

		if err := guildRepo.SetBotInstalled(c.Request().Context(), guildID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to mark bot as installed"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func verifyBotInstallHandler(guildRepo repositories.GuildRepository, botClient *discord.BotClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		ctx := c.Request().Context()

		guild, err := guildRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}
		if guild.OwnerDiscordID != claims.DiscordID {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "you do not own this guild"})
		}

		installed, err := botClient.VerifyBotInGuild(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]interface{}{"ok": false, "error": "failed to verify bot installation"})
		}
		if !installed {
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "error": "bot is not installed in this guild"})
		}

		discordRoles, err := botClient.GetGuildRoles(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]interface{}{"ok": false, "error": "failed to fetch guild roles"})
		}

		roles := make([]repositories.GuildRole, 0, len(discordRoles))
		for _, dr := range discordRoles {
			roles = append(roles, repositories.GuildRole{
				DiscordRoleID:  dr.ID,
				Position:       dr.Position,
				Type:           repositories.GuildRoleTypeDefault,
				AppPermissions: []string{},
				Managed:        dr.Managed,
				IsDefault:      dr.Name == "@everyone",
			})
		}

		if err := guildRepo.SetBotInstalledWithRoles(ctx, guildID, roles); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to mark bot as installed"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func memberSyncStatusHandler(guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		ctx := c.Request().Context()

		guild, err := guildRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}
		if guild.OwnerDiscordID != claims.DiscordID {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "you do not own this guild"})
		}

		members, err := memberRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok":          true,
			"memberCount": len(members),
			"synced":      len(members) > 0,
		})
	}
}

func syncMembersHandler(guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository, botClient *discord.BotClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		ctx := c.Request().Context()

		guild, err := guildRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}
		if guild.OwnerDiscordID != claims.DiscordID {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "you do not own this guild"})
		}

		// Derive filter and rank data from guild config once, before the member loop.
		activeRoleID := guild.NotificationConfig.StatusRoles.ActiveRoleID

		// Ranked roles are stored sorted by position descending (highest first) from UpsertRole.
		var rankedRoles []repositories.GuildRole
		for _, r := range guild.Roles {
			if r.Type == repositories.GuildRoleTypeRanked {
				rankedRoles = append(rankedRoles, r)
			}
		}

		discordMembers, err := botClient.GetGuildMembers(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]interface{}{"ok": false, "error": "failed to fetch guild members from Discord"})
		}

		// Track every Discord ID that passes the filter and gets upserted.
		syncedIDs := make(map[string]struct{}, len(discordMembers))

		synced := 0
		for _, dm := range discordMembers {
			if dm.User == nil || dm.User.ID == "" {
				continue
			}

			// O(1) role lookup set for this member.
			memberRoleSet := make(map[string]struct{}, len(dm.Roles))
			for _, rid := range dm.Roles {
				memberRoleSet[rid] = struct{}{}
			}

			// Skip unverified members when an active role is configured.
			if activeRoleID != "" {
				if _, has := memberRoleSet[activeRoleID]; !has {
					continue
				}
			}

			// Resolve highest-position ranked role the member holds.
			rankedRoleID := ""
			for _, rr := range rankedRoles {
				if _, has := memberRoleSet[rr.DiscordRoleID]; has {
					rankedRoleID = rr.DiscordRoleID
					break
				}
			}

			if err := memberRepo.Upsert(ctx, &repositories.Member{
				GuildID:         guildID,
				DiscordID:       dm.User.ID,
				RoleIDs:         dm.Roles,
				Status:          repositories.MemberStatusActive,
				RankedRoleID:    rankedRoleID,
				DiscordJoinedAt: dm.JoinedAt,
			}); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to upsert member"})
			}

			// Record that this member is confirmed present in Discord.
			syncedIDs[dm.User.ID] = struct{}{}
			synced++
		}

		// Reconcile: any member in our DB that Discord did not return is no
		// longer in the guild (or no longer passes the verified role filter).
		// Mark them inactive so the DB reflects the current guild state.
		storedMembers, err := memberRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		for _, m := range storedMembers {
			if _, stillPresent := syncedIDs[m.DiscordID]; !stillPresent {
				if m.Status == repositories.MemberStatusActive {
					if err := memberRepo.UpdateStatusAndRank(ctx, guildID, m.DiscordID, repositories.MemberStatusInactive, ""); err != nil {
						return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to deactivate departed member"})
					}
				}
			}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "synced": synced})
	}
}

func guildDashboardHandler(
	guildRepo repositories.GuildRepository,
	memberRepo repositories.MemberRepository,
	eventRepo repositories.EventRepository,
	eventReportRepo repositories.EventReportRepository,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		leaderboardBy := strings.ToLower(strings.TrimSpace(c.QueryParam("leaderboardBy")))
		if leaderboardBy == "" {
			leaderboardBy = "score"
		}
		if leaderboardBy != "score" && leaderboardBy != "hosted" && leaderboardBy != "attended" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "leaderboardBy must be score|hosted|attended"})
		}

		leaderboardLimit := 10
		if raw := strings.TrimSpace(c.QueryParam("leaderboardLimit")); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v <= 0 {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "leaderboardLimit must be a positive integer"})
			}
			if v > 200 {
				v = 200
			}
			leaderboardLimit = v
		}

		inactiveDays := 30
		if raw := strings.TrimSpace(c.QueryParam("inactiveDays")); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v <= 0 {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "inactiveDays must be a positive integer"})
			}
			inactiveDays = v
		}

		var eventStart *time.Time
		var eventEnd *time.Time
		if raw := strings.TrimSpace(c.QueryParam("eventStart")); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventStart must be RFC3339"})
			}
			eventStart = &t
		}
		if raw := strings.TrimSpace(c.QueryParam("eventEnd")); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventEnd must be RFC3339"})
			}
			eventEnd = &t
		}
		if eventStart != nil && eventEnd != nil && eventEnd.Before(*eventStart) {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventEnd must be >= eventStart"})
		}

		attendeeID := strings.TrimSpace(c.QueryParam("attendeeId"))
		// memberSearch intentionally left client-side for now per design.

		ctx := c.Request().Context()

		guild, err := guildRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}

		// One-shot activity map from event_reports (source of truth).
		activityList, err := eventReportRepo.GetGuildMemberActivity(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to load member activity"})
		}
		activityByDiscordID := make(map[string]repositories.GuildDashboardLeaderboardEntry, len(activityList))
		for _, entry := range activityList {
			activityByDiscordID[entry.DiscordID] = entry
		}

		memberRows, inactiveRows, err := memberRepo.GetDashboardMemberRows(ctx, guildID, inactiveDays)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to load dashboard members"})
		}

		memberCounts, err := memberRepo.GetGuildMemberCounts(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to load member counts"})
		}
		totalMembers := memberCounts.TotalMembers
		activeMembers := memberCounts.ActiveMembers
		inactiveMembers := memberCounts.InactiveMembers

		// liveEvents: count of open/active events from events collection
		liveEvents, err := eventRepo.GetLiveEventCounts(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to load event counts"})
		}

		// closedEvents: count from event_reports collection (source of truth for closed events)
		closedEvents, err := eventReportRepo.GetGuildClosedEventCount(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to load closed event count"})
		}

		// totalEvents = live events + closed events
		totalEvents := liveEvents + closedEvents

		// participationRate: percentage of active members who have attended at least one closed event.
		// Formula: (distinct members with at least one attended event) / activeMembers * 100
		// Clamped to [0, 100] and 0 when activeMembers is 0.
		membersWithAttendance := int64(0)
		for _, entry := range activityByDiscordID {
			if entry.AttendedCount > 0 {
				membersWithAttendance++
			}
		}

		participationRate := 0.0
		if activeMembers > 0 {
			participationRate = (float64(membersWithAttendance) / float64(activeMembers)) * 100.0
		}

		leaderboard := make([]repositories.GuildDashboardLeaderboardEntry, 0, len(activityByDiscordID))
		for _, entry := range activityByDiscordID {
			leaderboard = append(leaderboard, entry)
		}

		sort.SliceStable(leaderboard, func(i, j int) bool {
			li := leaderboard[i]
			lj := leaderboard[j]

			var primaryI int64
			var primaryJ int64
			switch leaderboardBy {
			case "hosted":
				primaryI = li.HostedCount
				primaryJ = lj.HostedCount
			case "attended":
				primaryI = li.AttendedCount
				primaryJ = lj.AttendedCount
			default:
				primaryI = li.Score
				primaryJ = lj.Score
			}

			if primaryI != primaryJ {
				return primaryI > primaryJ
			}
			return li.DiscordID < lj.DiscordID
		})

		if len(leaderboard) > leaderboardLimit {
			leaderboard = leaderboard[:leaderboardLimit]
		}
		for i := range leaderboard {
			leaderboard[i].Rank = int64(i + 1)
		}

		eventFilter := repositories.GuildDashboardEventFilter{
			StartDate:  eventStart,
			EndDate:    eventEnd,
			AttendeeID: attendeeID,
			Limit:      100,
		}

		events, err := eventReportRepo.FindDashboardEvents(ctx, guildID, eventFilter)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to load dashboard events"})
		}
		if events == nil {
			events = []repositories.GuildDashboardEventRow{}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok": true,
			"dashboard": repositories.GuildDashboardData{
				Guild: guild,
				Stats: repositories.GuildDashboardStats{
					TotalMembers:      totalMembers,
					ActiveMembers:     activeMembers,
					InactiveMembers:   inactiveMembers,
					TotalEvents:       totalEvents,
					ClosedEvents:      closedEvents,
					ParticipationRate: participationRate,
				},
				Leaderboard:     leaderboard,
				Members:         memberRows,
				InactiveMembers: inactiveRows,
				Events:          events,
			},
		})
	}
}

func listDiscordGuildsHandler(userRepo repositories.UserRepository, oauthClient *discord.OAuthClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		user, err := userRepo.FindByDiscordID(c.Request().Context(), claims.DiscordID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if user == nil || user.AccessToken == "" {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "no Discord access token"})
		}

		guilds, err := oauthClient.GetUserGuilds(c.Request().Context(), user.AccessToken)
		if err != nil {
			return c.JSON(http.StatusBadGateway, map[string]interface{}{"ok": false, "error": "failed to fetch Discord guilds"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "guilds": guilds})
	}
}

func botInviteURLHandler(guildRepo repositories.GuildRepository, oauthClient *discord.OAuthClient) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		guild, err := guildRepo.FindByGuildID(c.Request().Context(), guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}
		if guild.OwnerDiscordID != claims.DiscordID {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "you do not own this guild"})
		}

		redirectURI := c.QueryParam("redirectUri")
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "url": oauthClient.BotInviteURL(guildID, redirectURI)})
	}
}

func upsertRoleHandler(guildRepo repositories.GuildRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		var in upsertRolePayload
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}
		if strings.TrimSpace(in.DiscordRoleID) == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "discordRoleId is required"})
		}
		if in.Type != repositories.GuildRoleTypeDefault && in.Type != repositories.GuildRoleTypeRanked {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "type must be 'default' or 'ranked'"})
		}

		ctx := c.Request().Context()

		guild, err := guildRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}
		if guild.OwnerDiscordID != claims.DiscordID {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "you do not own this guild"})
		}

		role := repositories.GuildRole{
			DiscordRoleID:  in.DiscordRoleID,
			Position:       in.Position,
			Type:           in.Type,
			AppPermissions: []string{},
			Managed:        in.Managed,
			IsDefault:      in.IsDefault,
		}

		if err := guildRepo.UpsertRole(ctx, guildID, role); err != nil {
			if errors.Is(err, repositories.ErrGuildNotFound) {
				return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to upsert role"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

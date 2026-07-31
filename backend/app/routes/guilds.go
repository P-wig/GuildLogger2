package routes

import (
	"context"
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

type guildListItem struct {
	repositories.Guild
	IsOwner bool `json:"isOwner"`
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
	g.GET("/guilds", listGuildsHandler(guildRepo, memberRepo))
	g.GET("/guilds/discord", listDiscordGuildsHandler(userRepo, oauthClient))
	g.POST("/guilds/connect", connectGuildHandler(guildRepo, userRepo, oauthClient))
	g.POST("/guilds/:guildId/bot/install", installBotHandler(guildRepo))
	g.POST("/guilds/:guildId/bot/verify", verifyBotInstallHandler(guildRepo, botClient))
	g.GET("/guilds/:guildId/members/sync-status", memberSyncStatusHandler(guildRepo, memberRepo))
	g.POST("/guilds/:guildId/members/sync", syncMembersHandler(guildRepo, memberRepo, botClient))
	g.GET("/guilds/:guildId/dashboard", guildDashboardHandler(guildRepo, memberRepo, eventRepo, eventReportRepo))
	g.GET("/guilds/:guildId/bot/invite-url", botInviteURLHandler(guildRepo, oauthClient))
	g.PUT("/guilds/:guildId/roles", upsertRoleHandler(guildRepo))
	g.PUT("/guilds/:guildId/config/member-role", updateMemberRoleHandler(guildRepo))
	g.PUT("/guilds/:guildId/config", updateGuildConfigHandler(guildRepo))
	g.PUT("/guilds/:guildId/config/event", updateEventConfigHandler(guildRepo))
	g.DELETE("/guilds/:guildId", deleteGuildHandler(guildRepo, memberRepo))
	g.GET("/guilds/:guildId/event-logs", listGuildEventLogsHandler(guildRepo, memberRepo, eventReportRepo))
	g.POST("/guilds/:guildId/event-logs", createGuildEventLogHandler(guildRepo, memberRepo, eventReportRepo))
	g.PUT("/guilds/:guildId/event-logs/:logId", updateGuildEventLogHandler(guildRepo, memberRepo, eventReportRepo))
	g.DELETE("/guilds/:guildId/event-logs/:logId", deleteGuildEventLogHandler(guildRepo, memberRepo, eventReportRepo))
}

func listGuildsHandler(guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		ctx := c.Request().Context()

		ownedGuilds, err := guildRepo.FindByOwnerDiscordID(ctx, claims.DiscordID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}

		ownedIDs := make(map[string]struct{}, len(ownedGuilds))
		for _, g := range ownedGuilds {
			ownedIDs[g.GuildID] = struct{}{}
		}

		memberGuildIDs, err := memberRepo.FindGuildIDsByMemberDiscordID(ctx, claims.DiscordID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}

		unownedIDs := make([]string, 0, len(memberGuildIDs))
		for _, id := range memberGuildIDs {
			if _, isOwned := ownedIDs[id]; !isOwned {
				unownedIDs = append(unownedIDs, id)
			}
		}

		memberGuilds, err := guildRepo.FindByGuildIDs(ctx, unownedIDs)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}

		result := make([]guildListItem, 0, len(ownedGuilds)+len(memberGuilds))
		for _, g := range ownedGuilds {
			result = append(result, guildListItem{Guild: g, IsOwner: true})
		}
		for _, g := range memberGuilds {
			result = append(result, guildListItem{Guild: g, IsOwner: false})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "guilds": result})
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
				Name:           dr.Name,
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

		if guild.NotificationConfig.StatusRoles.ActiveRoleID == "" {
			return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{"ok": false, "error": "member role not configured: set a member role in the guild dashboard before syncing"})
		}

		// Derive filter and rank data from guild config once, before the member loop.
		activeRoleID := guild.NotificationConfig.StatusRoles.ActiveRoleID
		inactiveRoleID := guild.NotificationConfig.StatusRoles.InactiveRoleID

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

			// Determine guild membership status by role.
			// The active role is the baseline — anyone without it is not a guild member.
			// Members who also hold the inactive role are considered retired.
			if _, hasActive := memberRoleSet[activeRoleID]; !hasActive {
				continue
			}

			memberStatus := repositories.MemberStatusActive
			if inactiveRoleID != "" {
				if _, hasInactive := memberRoleSet[inactiveRoleID]; hasInactive {
					memberStatus = repositories.MemberStatusInactive
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

			// Prefer server nick > global display name > username for display.
			displayName := dm.User.Username
			if dm.User.GlobalName != "" {
				displayName = dm.User.GlobalName
			}
			if dm.Nick != "" {
				displayName = dm.Nick
			}

			if err := memberRepo.Upsert(ctx, &repositories.Member{
				GuildID:         guildID,
				DiscordID:       dm.User.ID,
				Username:        displayName,
				AvatarHash:      dm.User.Avatar,
				RoleIDs:         dm.Roles,
				Status:          memberStatus,
				RankedRoleID:    rankedRoleID,
				DiscordJoinedAt: dm.JoinedAt,
			}); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to upsert member"})
			}

			// Record that this member is confirmed present in Discord.
			syncedIDs[dm.User.ID] = struct{}{}
			synced++
		}

		// Reconcile: any member in our DB that Discord did not return (or who no
		// longer holds a qualifying role) is no longer a guild member — remove them.
		storedMembers, err := memberRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		for _, m := range storedMembers {
			if _, stillPresent := syncedIDs[m.DiscordID]; !stillPresent {
				if err := memberRepo.Delete(ctx, guildID, m.DiscordID); err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to remove departed member"})
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

		// Enrich member rows with event stats from activityByDiscordID.
		for i, row := range memberRows {
			if activity, ok := activityByDiscordID[row.DiscordID]; ok {
				memberRows[i].EventsHosted = activity.HostedCount
				memberRows[i].EventsAttended = activity.AttendedCount
				memberRows[i].LastHostedDate = activity.LastHostedDate
				memberRows[i].LastAttendedDate = activity.LastAttendedDate
			}
		}
		for i, row := range inactiveRows {
			if activity, ok := activityByDiscordID[row.DiscordID]; ok {
				inactiveRows[i].LastHostedDate = activity.LastHostedDate
				inactiveRows[i].LastAttendedDate = activity.LastAttendedDate
				if activity.LastHostedDate != nil || activity.LastAttendedDate != nil {
					var latest *time.Time
					if activity.LastHostedDate != nil {
						latest = activity.LastHostedDate
					}
					if activity.LastAttendedDate != nil && (latest == nil || activity.LastAttendedDate.After(*latest)) {
						latest = activity.LastAttendedDate
					}
					inactiveRows[i].LastActivityDate = latest
					days := int64(time.Since(*latest).Hours() / 24)
					inactiveRows[i].DaysSinceActivity = &days
				}
			}
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
					TotalMembers:    totalMembers,
					ActiveMembers:   activeMembers,
					InactiveMembers: inactiveMembers, LiveEvents: liveEvents, TotalEvents: totalEvents,
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

func updateGuildConfigHandler(guildRepo repositories.GuildRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		var in struct {
			ActiveRoleID     string   `json:"activeRoleId"`
			InactiveRoleID   string   `json:"inactiveRoleId"`
			RankedRoleIDs    []string `json:"rankedRoleIds"`
			ModeratorRoleIDs []string `json:"moderatorRoleIds"`
		}
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}
		in.ActiveRoleID = strings.TrimSpace(in.ActiveRoleID)
		if in.ActiveRoleID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "activeRoleId is required"})
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

		roleSet := make(map[string]struct{}, len(guild.Roles))
		for _, r := range guild.Roles {
			roleSet[r.DiscordRoleID] = struct{}{}
		}
		if _, ok := roleSet[in.ActiveRoleID]; !ok {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "activeRoleId not found in guild roles"})
		}
		in.InactiveRoleID = strings.TrimSpace(in.InactiveRoleID)
		if in.InactiveRoleID != "" {
			if _, ok := roleSet[in.InactiveRoleID]; !ok {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "inactiveRoleId not found in guild roles"})
			}
		}

		cleanModeratorRoleIDs := make([]string, 0, len(in.ModeratorRoleIDs))
		for _, modRoleID := range in.ModeratorRoleIDs {
			trimmed := strings.TrimSpace(modRoleID)
			if trimmed == "" {
				continue
			}
			if _, ok := roleSet[trimmed]; !ok {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "moderatorRoleId not found in guild roles: " + trimmed})
			}
			cleanModeratorRoleIDs = append(cleanModeratorRoleIDs, trimmed)
		}

		rankedSet := make(map[string]struct{}, len(in.RankedRoleIDs))
		for _, id := range in.RankedRoleIDs {
			rankedSet[strings.TrimSpace(id)] = struct{}{}
		}
		for i, r := range guild.Roles {
			if _, isRanked := rankedSet[r.DiscordRoleID]; isRanked {
				guild.Roles[i].Type = repositories.GuildRoleTypeRanked
			} else {
				guild.Roles[i].Type = repositories.GuildRoleTypeDefault
			}
		}

		cfg := repositories.GuildStatusRoleConfig{
			ActiveRoleID:     in.ActiveRoleID,
			InactiveRoleID:   in.InactiveRoleID,
			ModeratorRoleIDs: cleanModeratorRoleIDs,
		}
		if err := guildRepo.UpdateConfigAndRoles(ctx, guildID, cfg, guild.Roles); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to save guild config"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func updateEventConfigHandler(guildRepo repositories.GuildRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}
		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}
		var in struct {
			EventsChannelID string   `json:"eventsChannelId"`
			EventTypes      []string `json:"eventTypes"`
		}
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
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
		cleanTypes := make([]string, 0, len(in.EventTypes))
		for _, t := range in.EventTypes {
			if t = strings.TrimSpace(t); t != "" {
				cleanTypes = append(cleanTypes, t)
			}
		}
		cfg := repositories.GuildEventConfig{
			EventsChannelID: strings.TrimSpace(in.EventsChannelID),
			EventTypes:      cleanTypes,
		}
		if err := guildRepo.UpdateEventConfig(ctx, guildID, cfg); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to save event config"})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func updateMemberRoleHandler(guildRepo repositories.GuildRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		var in struct {
			RoleID string `json:"roleId"`
		}
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}
		in.RoleID = strings.TrimSpace(in.RoleID)
		if in.RoleID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "roleId is required"})
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

		valid := false
		for _, r := range guild.Roles {
			if r.DiscordRoleID == in.RoleID {
				valid = true
				break
			}
		}
		if !valid {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "role not found in this guild"})
		}

		cfg := repositories.GuildStatusRoleConfig{
			ActiveRoleID:   in.RoleID,
			InactiveRoleID: guild.NotificationConfig.StatusRoles.InactiveRoleID,
		}
		if err := guildRepo.UpdateStatusRoleConfig(ctx, guildID, cfg); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to update member role"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

// memberTier represents the access level of a Discord user within a guild.
type memberTier int

const (
	tierNone      memberTier = 0 // not a guild member
	tierMember    memberTier = 1 // synced active member — read-only access
	tierModerator memberTier = 2 // holds a configured moderator role — can write event logs
	tierOwner     memberTier = 3 // guild owner — full access
)

// getGuildMemberTier returns the caller's access tier for a guild.
// Owner → tierOwner; moderator-role holder → tierModerator;
// any other synced member → tierMember; not a member → tierNone.
func getGuildMemberTier(ctx context.Context, memberRepo repositories.MemberRepository, guildID string, guild *repositories.Guild, discordID string) memberTier {
	if guild.OwnerDiscordID == discordID {
		return tierOwner
	}
	m, err := memberRepo.FindByGuildAndDiscordID(ctx, guildID, discordID)
	if err != nil || m == nil {
		return tierNone
	}
	modRoleIDs := guild.NotificationConfig.StatusRoles.ModeratorRoleIDs
	if len(modRoleIDs) > 0 {
		modSet := make(map[string]struct{}, len(modRoleIDs))
		for _, id := range modRoleIDs {
			modSet[id] = struct{}{}
		}
		for _, roleID := range m.RoleIDs {
			if _, ok := modSet[roleID]; ok {
				return tierModerator
			}
		}
	}
	return tierMember
}

func deleteGuildHandler(guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository) echo.HandlerFunc {
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

		if err := memberRepo.DeleteAllByGuildID(ctx, guildID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to remove guild members"})
		}

		if err := guildRepo.Delete(ctx, guildID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to delete guild"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func listGuildEventLogsHandler(guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository, reportRepo repositories.EventReportRepository) echo.HandlerFunc {
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
		if getGuildMemberTier(ctx, memberRepo, guildID, guild, claims.DiscordID) < tierMember {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "access denied"})
		}

		reports, err := reportRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if reports == nil {
			reports = []repositories.EventReport{}
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "logs": reports})
	}
}

func createGuildEventLogHandler(guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository, reportRepo repositories.EventReportRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		if guildID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId is required"})
		}

		var in struct {
			Summary        string    `json:"summary"`
			EventDate      time.Time `json:"eventDate"`
			ParticipantIDs []string  `json:"participantIds"`
			HostDiscordID  string    `json:"hostDiscordId"`
		}
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}
		in.Summary = strings.TrimSpace(in.Summary)
		if in.Summary == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "summary is required"})
		}
		if in.EventDate.IsZero() {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventDate is required"})
		}

		hostID := strings.TrimSpace(in.HostDiscordID)
		if hostID == "" {
			hostID = claims.DiscordID
		}

		ctx := c.Request().Context()

		guild, err := guildRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}
		if getGuildMemberTier(ctx, memberRepo, guildID, guild, claims.DiscordID) < tierModerator {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "access denied"})
		}

		if in.ParticipantIDs == nil {
			in.ParticipantIDs = []string{}
		}

		report := &repositories.EventReport{
			EventID:              "",
			GuildID:              guildID,
			HostDiscordID:        hostID,
			EventDate:            in.EventDate,
			ParticipantIDs:       in.ParticipantIDs,
			Summary:              in.Summary,
			SubmittedByDiscordID: claims.DiscordID,
		}
		if err := reportRepo.Create(ctx, report); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to create event log"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true, "log": report})
	}
}

func updateGuildEventLogHandler(guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository, reportRepo repositories.EventReportRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		logID := strings.TrimSpace(c.Param("logId"))
		if guildID == "" || logID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId and logId are required"})
		}

		var in struct {
			Summary        string    `json:"summary"`
			EventDate      time.Time `json:"eventDate"`
			ParticipantIDs []string  `json:"participantIds"`
			HostDiscordID  string    `json:"hostDiscordId"`
		}
		if err := c.Bind(&in); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "invalid request body"})
		}
		in.Summary = strings.TrimSpace(in.Summary)
		if in.Summary == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "summary is required"})
		}
		if in.EventDate.IsZero() {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "eventDate is required"})
		}

		ctx := c.Request().Context()

		guild, err := guildRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}
		if getGuildMemberTier(ctx, memberRepo, guildID, guild, claims.DiscordID) < tierModerator {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "access denied"})
		}

		hostID := strings.TrimSpace(in.HostDiscordID)
		if hostID == "" {
			hostID = claims.DiscordID
		}
		if in.ParticipantIDs == nil {
			in.ParticipantIDs = []string{}
		}

		report := &repositories.EventReport{
			HostDiscordID:  hostID,
			EventDate:      in.EventDate,
			ParticipantIDs: in.ParticipantIDs,
			Summary:        in.Summary,
		}
		if err := reportRepo.Update(ctx, logID, report); err != nil {
			if errors.Is(err, repositories.ErrReportNotFound) {
				return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event log not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to update event log"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

func deleteGuildEventLogHandler(guildRepo repositories.GuildRepository, memberRepo repositories.MemberRepository, reportRepo repositories.EventReportRepository) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, ok := c.Get("user").(*session.Claims)
		if !ok || claims == nil {
			return c.JSON(http.StatusUnauthorized, map[string]interface{}{"ok": false, "error": "missing session"})
		}

		guildID := strings.TrimSpace(c.Param("guildId"))
		logID := strings.TrimSpace(c.Param("logId"))
		if guildID == "" || logID == "" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"ok": false, "error": "guildId and logId are required"})
		}

		ctx := c.Request().Context()

		guild, err := guildRepo.FindByGuildID(ctx, guildID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "database error"})
		}
		if guild == nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "guild not found"})
		}
		if getGuildMemberTier(ctx, memberRepo, guildID, guild, claims.DiscordID) < tierModerator {
			return c.JSON(http.StatusForbidden, map[string]interface{}{"ok": false, "error": "access denied"})
		}

		if err := reportRepo.Delete(ctx, logID); err != nil {
			if errors.Is(err, repositories.ErrReportNotFound) {
				return c.JSON(http.StatusNotFound, map[string]interface{}{"ok": false, "error": "event log not found"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "failed to delete event log"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"ok": true})
	}
}

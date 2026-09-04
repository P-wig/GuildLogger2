package discord

import (
	"context"
	"errors"
	"time"

	"github.com/P-wig/GuildLogger2/backend/app/repositories"
)

var ErrMemberNotFound = errors.New("member not found in this guild")

// MemberProfile is the composed view of a member's standing in a guild.
type MemberProfile struct {
	DiscordID         string                    `json:"discordId"`
	Username          string                    `json:"username"`
	AvatarHash        string                    `json:"avatarHash,omitempty"`
	Status            repositories.MemberStatus `json:"status"`
	RankRoleID        string                    `json:"rankRoleId,omitempty"`
	RankName          string                    `json:"rankName,omitempty"`
	HostedCount       int64                     `json:"hostedCount"`
	ParticipatedCount int64                     `json:"participatedCount"`
	DiscordJoinedAt   time.Time                 `json:"discordJoinedAt"`
	FirstSyncedAt     time.Time                 `json:"firstSyncedAt"`
	DeactivatedAt     *time.Time                `json:"deactivatedAt,omitempty"`
}

// TenureDays returns whole days since the member joined the Discord guild.
func (p *MemberProfile) TenureDays() int {
	if p.DiscordJoinedAt.IsZero() {
		return 0
	}
	return int(time.Since(p.DiscordJoinedAt).Hours() / 24)
}

// StatsService composes member activity profiles from the member record and the permanent
// event-report history. It is a read model, not a lifecycle service: no writes and no
// Discord side effects. It exists so the bot and the REST API cannot disagree about what
// "hosted", "attended", or "rank" mean; each transport handles only presentation.
type StatsService struct {
	memberRepo repositories.MemberRepository
	guildRepo  repositories.GuildRepository
}

func NewStatsService(memberRepo repositories.MemberRepository, guildRepo repositories.GuildRepository) *StatsService {
	return &StatsService{memberRepo: memberRepo, guildRepo: guildRepo}
}

// MemberProfile builds a member's activity profile.
// Returns ErrMemberNotFound when the member has not been synced into the guild.
func (s *StatsService) MemberProfile(ctx context.Context, guildID, discordID string) (*MemberProfile, error) {
	member, err := s.memberRepo.FindByGuildAndDiscordID(ctx, guildID, discordID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrMemberNotFound
	}

	counts, err := s.memberRepo.GetStats(ctx, guildID, discordID)
	if err != nil {
		return nil, err
	}
	if counts == nil {
		return nil, ErrMemberNotFound
	}

	profile := &MemberProfile{
		DiscordID:         member.DiscordID,
		Username:          member.Username,
		AvatarHash:        member.AvatarHash,
		Status:            member.Status,
		RankRoleID:        member.RankedRoleID,
		HostedCount:       counts.HostedCount,
		ParticipatedCount: counts.ParticipatedCount,
		DiscordJoinedAt:   counts.DiscordJoinedAt,
		FirstSyncedAt:     counts.FirstSyncedAt,
		DeactivatedAt:     counts.DeactivatedAt,
	}

	// Rank is stored as a role ID; the display name lives on the guild's mirrored role list.
	if member.RankedRoleID != "" {
		if guild, gErr := s.guildRepo.FindByGuildID(ctx, guildID); gErr == nil && guild != nil {
			for _, role := range guild.Roles {
				if role.DiscordRoleID == member.RankedRoleID {
					profile.RankName = role.Name
					break
				}
			}
		}
	}

	return profile, nil
}

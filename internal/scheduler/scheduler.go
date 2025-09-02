package scheduler

import (
	"context"
	"log"
	"time"

	"cucats.org/discord/internal/cam"
	"cucats.org/discord/internal/config"
	"cucats.org/discord/internal/database"
	"cucats.org/discord/internal/discord"
	"github.com/bwmarrin/discordgo"
)

type Scheduler struct {
	db *database.DB
}

func New(db *database.DB) *Scheduler {
	return &Scheduler{db: db}
}

func (s *Scheduler) PeriodicUpdate() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.updateAllUsers()
		log.Println("Completed periodic token refresh and role update")
	}
}

func (s *Scheduler) updateAllUsers() {
	ctx := context.Background()

	users, err := s.db.GetAllUserTokens()
	if err != nil {
		log.Printf("Failed to get users: %v", err)
		return
	}

	for _, user := range users {
		discordToken, err := s.db.GetDiscordToken(user.DiscordUserID)

		camToken, err := s.db.GetEntraToken(user.DiscordUserID)

		// Get updated user info from Microsoft
		msUserInfo, err := cam.GetUserInfo(ctx, camToken.AccessToken)
		if err != nil {
			log.Printf("Failed to get MS user info for user %s: %v", user.DiscordUserID, err)
			continue
		}

		// Update Discord role connection
		roleConnection := &discordgo.ApplicationRoleConnection{
			PlatformName:     "University of Cambridge",
			PlatformUsername: user.EntraUPN,
			Metadata: map[string]string{
				"is_student": discord.BoolToString(msUserInfo.IsStudent),
				"is_staff":   discord.BoolToString(msUserInfo.IsStaff),
				"is_alumni":  discord.BoolToString(msUserInfo.IsAlumni),
				"college":    discord.IntToString(int(msUserInfo.College)),
			},
		}

		discordSession, _ := discordgo.New("Bearer " + discordToken.AccessToken)

		if _, err := discordSession.UserApplicationRoleConnectionUpdate(config.DiscordClientID, roleConnection); err != nil {
			log.Printf("Failed to update Discord role for user %s: %v", user.DiscordUserID, err)
		}
	}
}

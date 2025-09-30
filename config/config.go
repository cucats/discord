package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

var (
	Host                string
	DiscordInviteURL    string
	DiscordBotToken     string
	DiscordClientID     string
	DiscordClientSecret string
	CamClientID         string
	CamClientSecret     string
)

const GuildID = "785990750042980352"

func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		slog.Error("Required environment variable not set", "key", key)
		os.Exit(1)
	}
	return value
}

func Init() {
	godotenv.Load()

	Host = mustGetEnv("HOST")
	DiscordInviteURL = mustGetEnv("DISCORD_INVITE_URL")
	DiscordBotToken = mustGetEnv("DISCORD_BOT_TOKEN")
	DiscordClientID = mustGetEnv("DISCORD_CLIENT_ID")
	DiscordClientSecret = mustGetEnv("DISCORD_CLIENT_SECRET")
	CamClientID = mustGetEnv("CAM_CLIENT_ID")
	CamClientSecret = mustGetEnv("CAM_CLIENT_SECRET")
}

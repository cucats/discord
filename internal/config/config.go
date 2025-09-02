package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	DiscordCallbackPath = "/discord/callback"
	CamCallbackPath     = "/cam/callback"
)

var (
	Host                string
	DiscordInviteURL    string
	DiscordBotToken     string
	DiscordClientID     string
	DiscordClientSecret string
	DiscordRedirectURI  string
	CamClientID         string
	CamClientSecret     string
	CamRedirectURI      string
	DatabaseURL         string
)

func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Required environment variable %s not set", key)
	}
	return value
}

func init() {
	godotenv.Load()

	Host = mustGetEnv("HOST")
	DiscordInviteURL = mustGetEnv("DISCORD_INVITE_URL")
	DiscordBotToken = mustGetEnv("DISCORD_BOT_TOKEN")
	DiscordClientID = mustGetEnv("DISCORD_CLIENT_ID")
	DiscordClientSecret = mustGetEnv("DISCORD_CLIENT_SECRET")
	DiscordRedirectURI = Host + DiscordCallbackPath
	CamClientID = mustGetEnv("CAM_CLIENT_ID")
	CamClientSecret = mustGetEnv("CAM_CLIENT_SECRET")
	CamRedirectURI = Host + CamCallbackPath
	DatabaseURL = mustGetEnv("DATABASE_URL")
}

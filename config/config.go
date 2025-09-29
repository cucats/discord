package config

import (
	"log"
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
	DatabaseURL         string
)

func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Required environment variable %s not set", key)
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
	DatabaseURL = mustGetEnv("DATABASE_URL")
}

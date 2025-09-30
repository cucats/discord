package role

import (
	"cucats.org/discord/config"
	"golang.org/x/oauth2"
)

const discordAuthority = "https://discord.com/api/v10"

var DiscordOAuth *oauth2.Config

func InitDiscordOAuth() {
	DiscordOAuth = &oauth2.Config{
		ClientID:     config.DiscordClientID,
		ClientSecret: config.DiscordClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  discordAuthority + "/oauth2/authorize",
			TokenURL: discordAuthority + "/oauth2/token",
		},
		RedirectURL: config.Host + "/discord/callback",
		Scopes:      []string{"identify"},
	}
}

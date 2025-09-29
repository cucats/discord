package role

import (
	"fmt"
	"log"

	"cucats.org/discord/config"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/oauth2"
)

const discordAuthority = "https://discord.com/api/v10"

func BoolToString(b bool) string {
	if b {
		return "1"
	} else {
		return "0"
	}
}

func IntToString(i int) string {
	return fmt.Sprintf("%d", i)
}

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
		Scopes:      []string{"identify", "role_connections.write"},
	}
}

func RegisterMetadata() {
	session, _ := discordgo.New("Bot " + config.DiscordBotToken)

	metadata := []*discordgo.ApplicationRoleConnectionMetadata{
		{
			Type:        discordgo.ApplicationRoleConnectionMetadataBooleanEqual,
			Key:         "is_student",
			Name:        "Current student",
			Description: "Currently enrolled as a student",
		},
		{
			Type:        discordgo.ApplicationRoleConnectionMetadataBooleanEqual,
			Key:         "is_staff",
			Name:        "Staff member",
			Description: "University staff member",
		},
		{
			Type:        discordgo.ApplicationRoleConnectionMetadataBooleanEqual,
			Key:         "is_alumni",
			Name:        "Alumni",
			Description: "University alumni",
		},
		{
			Type:        discordgo.ApplicationRoleConnectionMetadataIntegerEqual,
			Key:         "college",
			Name:        "College",
			Description: "Member of a Cambridge college (0 = None, or integer from 1 to 31 for colleges ordered alphabetically)",
		},
	}

	_, err := session.ApplicationRoleConnectionMetadataUpdate(config.DiscordClientID, metadata)

	if err != nil {
		panic(err)
	}

	log.Println("Registered metadata")
}

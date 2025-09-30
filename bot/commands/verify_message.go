package commands

import (
	"cucats.org/discord/config"
	"github.com/bwmarrin/discordgo"
)

type verifyCommand struct{}

func init() {
	register(&verifyCommand{})
}

func (c *verifyCommand) definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:                     "verify_message",
		Description:              "Post verification instructions with link button",
		DefaultMemberPermissions: &[]int64{discordgo.PermissionAdministrator}[0],
	}
}

func (c *verifyCommand) handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title:       "Verify to access the server",
					Description: "Click the button below to verify your Cambridge account and access the server.",
					Color:       0x9B59B6,
				},
			},
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label: "Verify",
							Style: discordgo.LinkButton,
							URL:   config.Host + "/role",
						},
					},
				},
			},
		},
	})

	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to post verification message.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

package commands

import (
	"github.com/bwmarrin/discordgo"
)

type roleCommand struct{}

func init() {
	register(&roleCommand{})
}

func (c *roleCommand) definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "test",
		Description: "Test",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "test",
				Description: "Test",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "parameter",
						Description: "Description",
						Required:    true,
					},
				},
			},
		},
		DefaultMemberPermissions: &[]int64{discordgo.PermissionManageRoles}[0],
	}
}

func (c *roleCommand) handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options

	switch options[0].Name {
	}
}

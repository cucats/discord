package commands

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

type roleCommand struct{}

func init() {
	register(&roleCommand{})
}

func (c *roleCommand) definition() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "role",
		Description: "Manage roles for all members",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "add",
				Description: "Add a role to all members",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "The role to add",
						Required:    true,
					},
				},
			},
			{
				Name:        "remove",
				Description: "Remove a role from all members",
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionRole,
						Name:        "role",
						Description: "The role to remove",
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
	case "add":
		c.handleAdd(s, i, options[0].Options)
	case "remove":
		c.handleRemove(s, i, options[0].Options)
	}
}

func (c *roleCommand) handleAdd(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring response: %v", err)
		return
	}

	roleID := options[0].RoleValue(s, i.GuildID).ID

	var allMembers []*discordgo.Member
	after := ""
	for {
		members, err := s.GuildMembers(i.GuildID, after, 1000)
		if err != nil {
			log.Printf("Error fetching members: %v", err)
			message := "Error fetching server members."
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &message,
			})
			return
		}

		if len(members) == 0 {
			break
		}

		allMembers = append(allMembers, members...)
		after = members[len(members)-1].User.ID
	}

	successCount := 0

	for _, member := range allMembers {
		if member.User.Bot {
			continue
		}

		err := s.GuildMemberRoleAdd(i.GuildID, member.User.ID, roleID)
		if err == nil {
			successCount++
		}
	}

	message := fmt.Sprintf("Role added to %d members.", successCount)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
}

func (c *roleCommand) handleRemove(s *discordgo.Session, i *discordgo.InteractionCreate, options []*discordgo.ApplicationCommandInteractionDataOption) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		log.Printf("Error deferring response: %v", err)
		return
	}

	roleID := options[0].RoleValue(s, i.GuildID).ID

	var allMembers []*discordgo.Member
	after := ""
	for {
		members, err := s.GuildMembers(i.GuildID, after, 1000)
		if err != nil {
			log.Printf("Error fetching members: %v", err)
			message := "Error fetching server members."
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &message,
			})
			return
		}

		if len(members) == 0 {
			break
		}

		allMembers = append(allMembers, members...)
		after = members[len(members)-1].User.ID
	}

	successCount := 0

	for _, member := range allMembers {
		if member.User.Bot {
			continue
		}

		err := s.GuildMemberRoleRemove(i.GuildID, member.User.ID, roleID)
		if err == nil {
			successCount++
		}
	}

	message := fmt.Sprintf("Role removed from %d members.", successCount)

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &message,
	})
}

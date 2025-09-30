package commands

import (
	"github.com/bwmarrin/discordgo"
)

type handler interface {
	definition() *discordgo.ApplicationCommand
	handle(s *discordgo.Session, i *discordgo.InteractionCreate)
}

var handlers []handler

func register(h handler) {
	handlers = append(handlers, h)
}

func GetDefinitions() []*discordgo.ApplicationCommand {
	defs := make([]*discordgo.ApplicationCommand, len(handlers))
	for i, h := range handlers {
		defs[i] = h.definition()
	}
	return defs
}

func Handle(s *discordgo.Session, i *discordgo.InteractionCreate) {
	for _, h := range handlers {
		if h.definition().Name == i.ApplicationCommandData().Name {
			h.handle(s, i)
			return
		}
	}
}

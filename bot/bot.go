package bot

import (
	"log/slog"

	"cucats.org/discord/bot/commands"
	"cucats.org/discord/config"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	Session *discordgo.Session
}

func New() (*Bot, error) {
	session, err := discordgo.New("Bot " + config.DiscordBotToken)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		Session: session,
	}

	session.AddHandler(bot.ready)
	session.AddHandler(bot.interactionCreate)

	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers

	return bot, nil
}

func (b *Bot) Start() error {
	err := b.Session.Open()
	if err != nil {
		return err
	}
	slog.Info("bot started")
	return nil
}

func (b *Bot) Stop() error {
	return b.Session.Close()
}

func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	slog.Info("bot ready", "username", s.State.User.Username)

	commandDefs := commands.GetDefinitions()
	for _, cmd := range commandDefs {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			slog.Error("failed to register command", "command", cmd.Name, "error", err)
		} else {
			slog.Info("registered command", "command", cmd.Name)
		}
	}
}

func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		commands.Handle(s, i)
	}
}

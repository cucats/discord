package bot

import (
	"log"
	"strings"

	"cucats.org/discord/config"
	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session *discordgo.Session
}

func New() (*Bot, error) {
	session, err := discordgo.New("Bot " + config.DiscordBotToken)
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		session: session,
	}

	session.AddHandler(bot.messageCreate)
	session.AddHandler(bot.ready)

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	return bot, nil
}

func (b *Bot) Start() error {
	err := b.session.Open()
	if err != nil {
		return err
	}
	log.Println("Discord bot is now running")
	return nil
}

func (b *Bot) Stop() error {
	return b.session.Close()
}

func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Printf("Bot logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Simple ping command
	if strings.HasPrefix(m.Content, "!ping") {
		s.ChannelMessageSend(m.ChannelID, "Pong")
	}
}

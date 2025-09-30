package bot

import (
	"log"

	"cucats.org/discord/bot/commands"
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

	session.AddHandler(bot.ready)
	session.AddHandler(bot.interactionCreate)

	session.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMembers

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

	log.Println("Registering slash commands...")
	commandDefs := commands.GetDefinitions()

	for _, cmd := range commandDefs {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Error registering command %s: %v", cmd.Name, err)
		} else {
			log.Printf("Registered command: %s", cmd.Name)
		}
	}
}

func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		commands.Handle(s, i)
	}
}

package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"cucats.org/discord/bot"
	"cucats.org/discord/config"
	"cucats.org/discord/role"
	"cucats.org/discord/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	config.Init()
	role.InitDiscordOAuth()
	role.InitCamOAuth()

	discordBot, err := bot.New()
	if err != nil {
		slog.Error("failed to create bot", "error", err)
		os.Exit(1)
	}

	err = discordBot.Start()
	if err != nil {
		slog.Error("failed to start bot", "error", err)
		os.Exit(1)
	}
	defer discordBot.Stop()

	go func() {
		if err := server.Start(discordBot); err != nil {
			slog.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	slog.Info("shutting down")
}

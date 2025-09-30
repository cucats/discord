package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"cucats.org/discord/bot"
	"cucats.org/discord/config"
	"cucats.org/discord/role"
)

func main() {
	config.Init()
	role.InitDiscordOAuth()
	role.InitCamOAuth()

	go role.RegisterMetadata()
	go startHTTPServer()

	discordBot, err := bot.New()
	if err != nil {
		log.Fatal("Error creating Discord bot:", err)
	}

	err = discordBot.Start()
	if err != nil {
		log.Fatal("Error starting Discord bot:", err)
	}
	defer discordBot.Stop()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	log.Println("Shutting down gracefully...")
}

func startHTTPServer() {
	h := role.New()

	http.HandleFunc("/", h.Index)
	http.HandleFunc("/role", h.LinkedRole)
	http.HandleFunc("/discord/callback", h.DiscordCallback)
	http.HandleFunc("/cam/callback", h.CamCallback)

	port := "8080"
	log.Printf("HTTP server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("HTTP server error:", err)
	}
}

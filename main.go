package main

import (
	"log"
	"net/http"

	"cucats.org/discord/internal/config"
	"cucats.org/discord/internal/database"
	"cucats.org/discord/internal/discord"
	"cucats.org/discord/internal/handlers"
	"cucats.org/discord/internal/scheduler"
)

func main() {
	db, err := database.New(config.DatabaseURL)
	if err != nil {
		panic(err)
	}

	h := handlers.New(db)
	s := scheduler.New(db)

	go discord.RegisterMetadata()
	go s.PeriodicUpdate()

	http.HandleFunc("/", h.Index)
	http.HandleFunc("/linked-role", h.LinkedRole)
	http.HandleFunc(config.DiscordCallbackPath, h.DiscordCallback)
	http.HandleFunc(config.CamCallbackPath, h.CamCallback)

	port := "8080"
	log.Printf("Server starting on port %s", port)
	http.ListenAndServe(":"+port, nil)
}

package server

import (
	"log/slog"
	"net/http"

	"cucats.org/discord/bot"
	"cucats.org/discord/role"
)

func Start(discordBot *bot.Bot) error {
	mux := http.NewServeMux()
	h := role.New(discordBot)

	handle(mux, "/", h.Index)
	handle(mux, "/role", h.Role)
	handle(mux, "/discord/callback", h.DiscordCallback)
	handle(mux, "/cam/callback", h.CamCallback)

	port := "8080"
	slog.Info("http server starting", "port", port)
	return http.ListenAndServe(":"+port, mux)
}

func handle(mux *http.ServeMux, pattern string, handler http.HandlerFunc) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "ip", r.RemoteAddr)
		handler(w, r)
	})
}

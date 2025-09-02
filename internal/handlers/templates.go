package handlers

import (
	"html/template"
	"net/http"

	"cucats.org/discord/internal/cam"
	"github.com/bwmarrin/discordgo"
)

var templates *template.Template

func init() {
	var err error
	templates, err = template.ParseGlob("templates/*.gohtml")
	if err != nil {
		panic(err)
	}
}

type ErrorData struct {
	Message    string
	StatusCode int
}

func renderError(w http.ResponseWriter, message string, statusCode int) {
	data := ErrorData{
		Message:    message,
		StatusCode: statusCode,
	}

	templates.ExecuteTemplate(w, "error.gohtml", data)
}

type SuccessData struct {
	DiscordUsername string
	DiscordID       string
	UPN             string
	IsStudent       bool
}

func renderSuccess(w http.ResponseWriter, user *discordgo.User, msUserInfo *cam.UserInfo) {
	data := SuccessData{
		DiscordUsername: user.Username,
		DiscordID:       user.ID,
		UPN:             msUserInfo.UPN,
		IsStudent:       msUserInfo.IsStudent,
	}

	templates.ExecuteTemplate(w, "success.gohtml", data)
}

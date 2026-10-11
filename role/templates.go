package role

import (
	"html/template"
	"net/http"

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
	Account         string
	IsStudent       bool
}

func renderSuccess(w http.ResponseWriter, user *discordgo.User, msUserInfo *UserInfo) {
	data := SuccessData{
		DiscordUsername: user.Username,
		DiscordID:       user.ID,
		Account:         msUserInfo.Account,
		IsStudent:       msUserInfo.IsStudent,
	}

	templates.ExecuteTemplate(w, "success.gohtml", data)
}

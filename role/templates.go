package role

import (
	"html/template"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

var errorTemplate *template.Template
var successTemplate *template.Template

func init() {
	var err error
	errorTemplate, err = template.ParseFiles("templates/base.gohtml", "templates/error.gohtml")
	if err != nil {
		panic(err)
	}
	successTemplate, err = template.ParseFiles("templates/base.gohtml", "templates/success.gohtml")
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

	w.WriteHeader(statusCode)
	errorTemplate.ExecuteTemplate(w, "error.gohtml", data)
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

	successTemplate.ExecuteTemplate(w, "success.gohtml", data)
}

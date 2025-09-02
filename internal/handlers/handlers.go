package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"sync"

	"cucats.org/discord/internal/cam"
	"cucats.org/discord/internal/config"
	"cucats.org/discord/internal/database"
	"cucats.org/discord/internal/discord"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/oauth2"
)

type Handlers struct {
	db         *database.DB
	sessions   map[string]*Session
	sessionsMu sync.RWMutex
}

func New(db *database.DB) *Handlers {
	h := &Handlers{
		db:       db,
		sessions: make(map[string]*Session),
	}

	go h.sessionCleanup()

	return h
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GET /
func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, config.DiscordInviteURL, http.StatusFound)
}

// GET /linked-role
func (h *Handlers) LinkedRole(w http.ResponseWriter, r *http.Request) {
	session := h.createSession(w)
	session.DiscordState = generateState()

	authURL := discord.OAuth.AuthCodeURL(session.DiscordState)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// GET /discord/callback?state=<state>&code=<code>
func (h *Handlers) DiscordCallback(w http.ResponseWriter, r *http.Request) {
	session := h.getSession(r)
	if session == nil {
		renderError(w, "Session expired", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()

	if q.Get("state") != session.DiscordState {
		renderError(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := q.Get("code")
	if code == "" {
		renderError(w, "No authorization code", http.StatusBadRequest)
		return
	}

	token, err := discord.OAuth.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Discord token exchange error: %v", err)
		renderError(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	session.DiscordToken = token
	session.CamState = generateState()

	authURL := cam.OAuth.AuthCodeURL(session.CamState,
		oauth2.SetAuthURLParam("domain_hint", "cam.ac.uk"),
	)

	http.Redirect(w, r, authURL, http.StatusFound)
}

// GET /cam/callback?state=<state>&code=<code>
func (h *Handlers) CamCallback(w http.ResponseWriter, r *http.Request) {
	session := h.getSession(r)
	if session == nil || session.DiscordToken == nil {
		renderError(w, "Session expired", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()

	if q.Get("state") != session.CamState {
		renderError(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := q.Get("code")
	if code == "" {
		renderError(w, "No authorization code", http.StatusBadRequest)
		return
	}

	msToken, err := cam.OAuth.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("Microsoft token exchange error: %v", err)
		renderError(w, "Failed to exchange Microsoft code", http.StatusInternalServerError)
		return
	}

	discordSession, _ := discordgo.New("Bearer " + session.DiscordToken.AccessToken)
	discordUser, err := discordSession.User("@me")

	if err != nil {
		log.Printf("Discord user fetch error: %v", err)
		renderError(w, "Failed to get Discord user", http.StatusInternalServerError)
		return
	}

	msUser, err := cam.GetUserInfo(context.Background(), msToken.AccessToken)
	if err != nil {
		log.Printf("Microsoft user fetch error: %v", err)
		renderError(w, "Failed to get Microsoft user", http.StatusInternalServerError)
		return
	}

	// Save tokens to database
	userToken := &database.UserToken{
		DiscordUserID:       discordUser.ID,
		DiscordAccessToken:  session.DiscordToken.AccessToken,
		DiscordRefreshToken: session.DiscordToken.RefreshToken,
		EntraAccessToken:    msToken.AccessToken,
		EntraRefreshToken:   msToken.RefreshToken,
		EntraUPN:            msUser.UPN,
	}

	if err := h.db.SaveUserToken(userToken); err != nil {
		log.Printf("Database save error: %v", err)
		renderError(w, "Failed to save user data", http.StatusInternalServerError)
		return
	}

	// Update Discord role connection
	roleConnection := &discordgo.ApplicationRoleConnection{
		PlatformName:     "Cambridge Verification",
		PlatformUsername: msUser.UPN,
		Metadata: map[string]string{
			"is_student": discord.BoolToString(msUser.IsStudent),
			"is_staff":   discord.BoolToString(msUser.IsStaff),
			"is_alumni":  discord.BoolToString(msUser.IsAlumni),
			"college":    discord.IntToString(int(msUser.College)),
		},
	}

	_, err = discordSession.UserApplicationRoleConnectionUpdate(config.DiscordClientID, roleConnection)

	if err != nil {
		log.Printf("Discord role update error: %v", err)
		renderError(w, "Failed to update Discord role", http.StatusInternalServerError)
		return
	}

	renderSuccess(w, discordUser, msUser)
}

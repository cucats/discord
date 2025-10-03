package role

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
	"sync"

	"cucats.org/discord/bot"
	"cucats.org/discord/config"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/oauth2"
)

const verifiedRoleID = "1422372716840882348"

type Handlers struct {
	sessions   map[string]*Session
	sessionsMu sync.RWMutex
	bot        *bot.Bot
}

func New(discordBot *bot.Bot) *Handlers {
	h := &Handlers{
		sessions: make(map[string]*Session),
		bot:      discordBot,
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

// GET /role
func (h *Handlers) Role(w http.ResponseWriter, r *http.Request) {
	session := h.createSession(w)
	session.DiscordState = generateState()

	authURL := DiscordOAuth.AuthCodeURL(session.DiscordState)
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

	token, err := DiscordOAuth.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("discord token exchange failed", "error", err)
		renderError(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	session.DiscordToken = token
	session.CamState = generateState()

	authURL := CamOAuth.AuthCodeURL(session.CamState,
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

	msToken, err := CamOAuth.Exchange(context.Background(), code)
	if err != nil {
		slog.Error("cam token exchange failed", "error", err)
		renderError(w, "Failed to exchange Microsoft code", http.StatusInternalServerError)
		return
	}

	discordSession, _ := discordgo.New("Bearer " + session.DiscordToken.AccessToken)
	discordUser, err := discordSession.User("@me")
	if err != nil {
		slog.Error("discord user fetch failed", "error", err)
		renderError(w, "Failed to get Discord user", http.StatusInternalServerError)
		return
	}

	msUser, err := GetUserInfo(context.Background(), msToken.AccessToken)
	if err != nil {
		slog.Error("cam user fetch failed", "error", err)
		renderError(w, "Failed to get Microsoft user", http.StatusInternalServerError)
		return
	}

	slog.Info("user verified",
		"user_id", discordUser.ID,
		"username", discordUser.Username,
		"upn", msUser.UPN,
		"student", msUser.IsStudent,
		"staff", msUser.IsStaff,
		"alumni", msUser.IsAlumni,
		"college", msUser.College)

	err = h.bot.Session.GuildMemberRoleAdd(config.GuildID, discordUser.ID, verifiedRoleID)
	if err != nil {
		slog.Error("failed to add role", "user_id", discordUser.ID, "role_id", verifiedRoleID, "error", err)
		renderError(w, "failed to add verified role.", http.StatusInternalServerError)
		return
	}
	slog.Info("added role", "user_id", discordUser.ID, "role_id", verifiedRoleID)

	if msUser.College != Unknown {
		if collegeRoleID, ok := CollegeRoles[msUser.College]; ok {
			err = h.bot.Session.GuildMemberRoleAdd(config.GuildID, discordUser.ID, collegeRoleID)
			if err != nil {
				slog.Warn("failed to add role", "user_id", discordUser.ID, "role_id", collegeRoleID, "college", msUser.College, "error", err)
			} else {
				slog.Info("added role", "user_id", discordUser.ID, "role_id", collegeRoleID, "college", msUser.College)
			}
		}
	}

	renderSuccess(w, discordUser, msUser)
}

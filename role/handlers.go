package role

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"sync"

	"cucats.org/discord/config"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/oauth2"
)

type Handlers struct {
	sessions   map[string]*Session
	sessionsMu sync.RWMutex
}

func New() *Handlers {
	h := &Handlers{
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
		log.Printf("Discord token exchange error: %v", err)
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

	msUser, err := GetUserInfo(context.Background(), msToken.AccessToken)
	if err != nil {
		log.Printf("Microsoft user fetch error: %v", err)
		renderError(w, "Failed to get Microsoft user", http.StatusInternalServerError)
		return
	}

	log.Printf("User: Discord=%s#%s (ID=%s) UPN=%s Student=%v Staff=%v Alumni=%v College=%v",
		discordUser.Username, discordUser.Discriminator, discordUser.ID, msUser.UPN,
		msUser.IsStudent, msUser.IsStaff, msUser.IsAlumni, msUser.College)

	// Update Discord role connection
	roleConnection := &discordgo.ApplicationRoleConnection{
		PlatformName:     "Cambridge Verification",
		PlatformUsername: msUser.UPN,
		Metadata: map[string]string{
			"is_student": BoolToString(msUser.IsStudent),
			"is_staff":   BoolToString(msUser.IsStaff),
			"is_alumni":  BoolToString(msUser.IsAlumni),
			"college":    IntToString(int(msUser.College)),
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

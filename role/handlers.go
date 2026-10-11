package role

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
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

	token, err := DiscordOAuth.Exchange(r.Context(), code)
	if err != nil {
		slog.Error("discord token exchange failed", "error", err)
		renderError(w, "Failed to exchange code", http.StatusInternalServerError)
		return
	}

	session.DiscordToken = token
	h.redirectToKeycloakLogin(w, r, session, false)
}

func (h *Handlers) redirectToKeycloakLogin(w http.ResponseWriter, r *http.Request, session *Session, forceMicrosoft bool) {
	session.KeycloakState = generateState()
	session.KeycloakNonce = generateState()
	session.KeycloakVerifier = oauth2.GenerateVerifier()

	opts := []oauth2.AuthCodeOption{
		oauth2.S256ChallengeOption(session.KeycloakVerifier),
		oauth2.SetAuthURLParam("nonce", session.KeycloakNonce),
	}
	if forceMicrosoft {
		opts = append(opts, oauth2.SetAuthURLParam("prompt", "login"),
			oauth2.SetAuthURLParam("kc_idp_hint", config.KeycloakEntraAlias))
	}
	authURL := KeycloakOAuth.AuthCodeURL(session.KeycloakState, opts...)

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *Handlers) retryEntraLogin(w http.ResponseWriter, r *http.Request, session *Session) bool {
	if session.EntraRetry {
		return false
	}
	session.EntraRetry = true
	h.redirectToKeycloakLogin(w, r, session, true)
	return true
}

// GET /keycloak/callback?state=<state>&code=<code>
func (h *Handlers) KeycloakCallback(w http.ResponseWriter, r *http.Request) {
	session := h.getSession(r)
	if session == nil || session.DiscordToken == nil {
		renderError(w, "Session expired", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()

	if q.Get("state") != session.KeycloakState {
		renderError(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := q.Get("code")
	if code == "" {
		renderError(w, "No authorization code", http.StatusBadRequest)
		return
	}

	keycloakToken, err := KeycloakOAuth.Exchange(r.Context(), code,
		oauth2.VerifierOption(session.KeycloakVerifier))
	if err != nil {
		slog.Error("Keycloak token exchange failed", "error", err)
		renderError(w, "Failed to exchange Keycloak code", http.StatusInternalServerError)
		return
	}

	keycloakUser, err := GetKeycloakUserInfo(r.Context(), keycloakToken, session.KeycloakNonce)
	if err != nil {
		slog.Error("Keycloak user verification failed", "error", err)
		renderError(w, "Failed to verify Keycloak user", http.StatusInternalServerError)
		return
	}
	entraToken, err := GetBrokeredEntraToken(r.Context(), keycloakToken)
	if err != nil {
		slog.Error("Entra token retrieval failed", "error", err)
		if h.retryEntraLogin(w, r, session) {
			return
		}
		renderError(w, "Could not verify Cambridge account", http.StatusBadGateway)
		return
	}
	cambridgeUser, err := GetGraphUserInfo(r.Context(), entraToken)
	if err != nil {
		slog.Error("Cambridge group lookup failed", "error", err)
		if errors.Is(err, ErrGraphUnauthorized) && h.retryEntraLogin(w, r, session) {
			return
		}
		renderError(w, "Could not look up Cambridge groups", http.StatusBadGateway)
		return
	}
	if cambridgeUser.Account == "" {
		cambridgeUser.Account = keycloakUser.Account
	}
	if cambridgeUser.DisplayName == "" {
		cambridgeUser.DisplayName = keycloakUser.DisplayName
	}

	discordSession, _ := discordgo.New("Bearer " + session.DiscordToken.AccessToken)
	discordUser, err := discordSession.User("@me")
	if err != nil {
		slog.Error("discord user fetch failed", "error", err)
		renderError(w, "Failed to get Discord user", http.StatusInternalServerError)
		return
	}

	slog.Info("user verified",
		"user_id", discordUser.ID,
		"username", discordUser.Username,
		"account", cambridgeUser.Account,
		"student", cambridgeUser.IsStudent,
		"staff", cambridgeUser.IsStaff,
		"alumni", cambridgeUser.IsAlumni,
		"college", cambridgeUser.College)

	err = h.bot.Session.GuildMemberRoleAdd(config.GuildID, discordUser.ID, verifiedRoleID)
	if err != nil {
		slog.Error("failed to add role", "user_id", discordUser.ID, "role_id", verifiedRoleID, "error", err)
		renderError(w, "failed to add verified role.", http.StatusInternalServerError)
		return
	}
	slog.Info("added role", "user_id", discordUser.ID, "role_id", verifiedRoleID)

	if cambridgeUser.College != Unknown {
		if collegeRoleID, ok := CollegeRoles[cambridgeUser.College]; ok {
			err = h.bot.Session.GuildMemberRoleAdd(config.GuildID, discordUser.ID, collegeRoleID)
			if err != nil {
				slog.Warn("failed to add role", "user_id", discordUser.ID, "role_id", collegeRoleID, "college", cambridgeUser.College, "error", err)
			} else {
				slog.Info("added role", "user_id", discordUser.ID, "role_id", collegeRoleID, "college", cambridgeUser.College)
			}
		}
	}
	if err := StoreVerificationInKeycloak(r.Context(), keycloakUser.ID, discordUser.ID, discordUser.Username, cambridgeUser.College); err != nil {
		slog.Error("failed to store verification in Keycloak", "error", err)
		renderError(w, "Roles assigned, but could not save verification to Keycloak", http.StatusBadGateway)
		return
	}

	h.deleteSession(w, r)
	renderSuccess(w, discordUser, cambridgeUser)
}

func (h *Handlers) RobotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("User-agent: *\nDisallow: /\n"))
}

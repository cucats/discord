package handlers

import (
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type Session struct {
	DiscordState string
	DiscordToken *oauth2.Token
	CamState     string
	CreatedAt    time.Time
}

func (h *Handlers) sessionCleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.sessionsMu.Lock()
		now := time.Now()
		for sessionID, session := range h.sessions {
			if now.Sub(session.CreatedAt) > time.Hour {
				delete(h.sessions, sessionID)
			}
		}
		h.sessionsMu.Unlock()
	}
}

func (h *Handlers) createSession(w http.ResponseWriter) *Session {
	sessionID := generateState()
	session := &Session{CreatedAt: time.Now()}

	h.sessionsMu.Lock()
	h.sessions[sessionID] = session
	h.sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600, // 1 hour
	})

	return session
}

func (h *Handlers) getSession(r *http.Request) *Session {
	cookie, err := r.Cookie("session")
	if err != nil {
		return nil
	}

	h.sessionsMu.RLock()
	session, exists := h.sessions[cookie.Value]
	h.sessionsMu.RUnlock()

	if !exists {
		return nil
	}

	if time.Since(session.CreatedAt) > time.Hour {
		h.sessionsMu.Lock()
		delete(h.sessions, cookie.Value)
		h.sessionsMu.Unlock()
		return nil
	}

	return session
}

package database

const schema = `
CREATE TABLE IF NOT EXISTS user_token (
	id SERIAL PRIMARY KEY,
	discord_user_id VARCHAR(100) UNIQUE NOT NULL,
	discord_access_token TEXT NOT NULL,
	discord_refresh_token TEXT NOT NULL,
	entra_upn VARCHAR(200) NOT NULL,
	entra_access_token TEXT NOT NULL,
	entra_refresh_token TEXT NOT NULL,
	created_at TIMESTAMPTZ DEFAULT NOW(),
	updated_at TIMESTAMPTZ DEFAULT NOW()
);`

type UserToken struct {
	DiscordUserID       string `json:"discord_user_id"`
	DiscordAccessToken  string `json:"discord_access_token"`
	DiscordRefreshToken string `json:"discord_refresh_token"`
	EntraUPN            string `json:"entra_upn"`
	EntraAccessToken    string `json:"entra_access_token"`
	EntraRefreshToken   string `json:"entra_refresh_token"`
}

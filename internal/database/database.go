package database

import (
	"context"
	"database/sql"

	"cucats.org/discord/internal/cam"
	"cucats.org/discord/internal/discord"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"
)

type DB struct {
	*sql.DB
}

func New(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) SaveUserToken(token *UserToken) error {
	query := `
		INSERT INTO user_token (
			discord_user_id,
			discord_access_token,
			discord_refresh_token,
			entra_upn,
			entra_access_token,
			entra_refresh_token, 
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		ON CONFLICT (discord_user_id)
		DO UPDATE SET
			discord_access_token = EXCLUDED.discord_access_token,
			discord_refresh_token = EXCLUDED.discord_refresh_token,
			entra_upn = EXCLUDED.entra_upn,
			entra_access_token = EXCLUDED.entra_access_token,
			entra_refresh_token = EXCLUDED.entra_refresh_token,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := db.Exec(query,
		token.DiscordUserID,
		token.DiscordAccessToken,
		token.DiscordRefreshToken,
		token.EntraUPN,
		token.EntraAccessToken,
		token.EntraRefreshToken,
	)
	return err
}

func (db *DB) GetAllUserTokens() ([]*UserToken, error) {
	rows, err := db.Query(`
		SELECT 
			discord_user_id,
			discord_access_token,
			discord_refresh_token,
			entra_upn,
			entra_access_token,
			entra_refresh_token
		FROM user_token
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*UserToken
	for rows.Next() {
		var token UserToken

		err := rows.Scan(
			&token.DiscordUserID,
			&token.DiscordAccessToken,
			&token.DiscordRefreshToken,
			&token.EntraUPN,
			&token.EntraAccessToken,
			&token.EntraRefreshToken,
		)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, &token)
	}

	return tokens, rows.Err()
}

func (db *DB) GetDiscordToken(discordUserID string) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	err := db.QueryRow(`
		SELECT discord_access_token, discord_refresh_token
		FROM user_token
		WHERE discord_user_id = $1
	`, discordUserID).Scan(&token.AccessToken, &token.RefreshToken)
	if err != nil {
		return nil, err
	}

	tokenSource := discord.OAuth.TokenSource(context.Background(), token)

	token, err = tokenSource.Token()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		UPDATE user_token 
		SET discord_access_token = $2, discord_refresh_token = $3, updated_at = CURRENT_TIMESTAMP
		WHERE discord_user_id = $1
	`, discordUserID, token.AccessToken, token.RefreshToken)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (db *DB) GetEntraToken(discordUserID string) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	err := db.QueryRow(`
		SELECT entra_access_token, entra_refresh_token
		FROM user_token
		WHERE discord_user_id = $1
	`, discordUserID).Scan(&token.AccessToken, &token.RefreshToken)
	if err != nil {
		return nil, err
	}

	tokenSource := cam.OAuth.TokenSource(context.Background(), token)

	token, err = tokenSource.Token()
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		UPDATE user_token
		SET entra_access_token = $2, entra_refresh_token = $3, updated_at = CURRENT_TIMESTAMP
		WHERE discord_user_id = $1
	`, discordUserID, token.AccessToken, token.RefreshToken)
	if err != nil {
		return nil, err
	}

	return token, nil
}

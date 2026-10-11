<div align="center">

# Discord Linked Role for the University of Cambridge

[![CI](https://github.com/cucats/discord/actions/workflows/ci.yml/badge.svg)](https://github.com/cucats/discord/actions/workflows/ci.yml)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go](https://img.shields.io/badge/go-1.24.6%2B-blue)](https://golang.org/dl/)
[![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?logo=docker&logoColor=white)](https://hub.docker.com)

Verifies members through the society's Keycloak single sign-on for Discord roles.

</div>

## Prerequisites

- Go 1.24.6 or later
- Discord Application (Client ID, Client Secret, Bot Token) from https://discord.com/developers/applications
- A confidential OpenID Connect client in the `master` realm at https://kc.cucats.org

## Setup

Create environment file from template:

```sh
cp .env.example .env
```

Set these environment variables in `.env`:

```conf
HOST=http://localhost:8080

# Discord
DISCORD_INVITE_URL=https://discord.gg/your-invite
DISCORD_BOT_TOKEN=your_discord_bot_token
DISCORD_CLIENT_ID=your_discord_client_id
DISCORD_CLIENT_SECRET=your_discord_client_secret

# Keycloak
KEYCLOAK_ISSUER=https://kc.cucats.org/realms/master
KEYCLOAK_CLIENT_ID=your_keycloak_client_id
KEYCLOAK_CLIENT_SECRET=your_keycloak_client_secret
KEYCLOAK_WRITER_CLIENT_ID=cucats-discord-writer
KEYCLOAK_WRITER_CLIENT_SECRET=your_keycloak_writer_client_secret
KEYCLOAK_ENTRA_ALIAS=microsoft
```

Install dependencies:

```sh
go mod download
```

Run development server:

```sh
go run .
```

Build and run production binary:

```sh
go build -o discord
./discord
```

## Endpoints

- `/` - Redirects to Discord invite
- `/role` - Start verification process
- `/discord/callback` - Discord OAuth callback
- `/keycloak/callback` - Keycloak OIDC callback

Configure the Keycloak client with standard authorization code flow, client
authentication, and the valid redirect URI `${HOST}/keycloak/callback` (using the
public HTTPS URL in production). The flow uses PKCE and an ID token nonce. Users
with an existing Keycloak browser session can return without signing in again.

The app retrieves the linked Entra access token through Keycloak's broker API,
then calls Microsoft Graph `/me` and `/me/memberOf` to check Cambridge groups and
assign the verified and college roles. Keycloak's public login page currently
identifies the Microsoft provider as `microsoft`; set `KEYCLOAK_ENTRA_ALIAS` if
that alias changes. The app does not grant a verified role when it cannot fetch
the brokered token or Graph memberships.

After assigning roles, the bot stores `discord_id`, `discord_username`, and
`college` on the verified Keycloak user. It removes a previously stored college
when Cambridge no longer reports one. Create a separate confidential client with
service accounts enabled for `KEYCLOAK_WRITER_CLIENT_ID`, and grant its service
account permission to read and update realm users. Configure these three user
profile attributes with admin edit permission, so users cannot set their own
college or Discord identity. The bot reads the attributes back and reports an
error if Keycloak does not retain them.

The production Keycloak realm uses broker API V2. Enable it with
`KC_FEATURES=identity-brokering-api:v2`, set **Store token in session** on the
Microsoft identity provider, and allow this confidential client to retrieve
external tokens from the `microsoft` provider. The bot tries V2 first, then V1
for older realms. V1 requires **Store Tokens** and the `broker` client's
`read-token` role in both the user's role mappings and the OIDC client's scope.
See the [Keycloak identity brokering API documentation](https://www.keycloak.org/docs/latest/server_development/#_identity_brokering_apis).

Microsoft Graph requires a usable delegated `User.Read` token. A stored Entra
token may expire while a Keycloak browser session remains active; the broker
must be able to return a current token for college lookup to succeed. If no
usable token is available, the bot tries one fresh Microsoft broker login; an
existing Microsoft browser session may satisfy this without a password prompt.

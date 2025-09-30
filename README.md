<div align="center">

# Discord Linked Role for the University of Cambridge

[![CI](https://github.com/cucats/discord/actions/workflows/ci.yml/badge.svg)](https://github.com/cucats/discord/actions/workflows/ci.yml)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go](https://img.shields.io/badge/go-1.24.6%2B-blue)](https://golang.org/dl/)
[![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?logo=docker&logoColor=white)](https://hub.docker.com)

Verifies University of Cambridge students via Microsoft Entra ID for Discord linked roles.

</div>

## Prerequisites

- Go 1.24.6 or later
- Discord Application (Client ID, Client Secret, Bot Token) from https://discord.com/developers/applications
- Microsoft Entra ID Application (Client ID, Client Secret) from https://toolkit.uis.cam.ac.uk/endpoints

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

# Microsoft Entra ID
UCAM_CLIENT_ID=your_microsoft_client_id
UCAM_CLIENT_SECRET=your_microsoft_client_secret
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

## Metadata

- `is_student`: Current student (boolean)
- `is_staff`: Staff member (boolean)
- `is_alumni`: Alumni (boolean)
- `college`: College membership (integer, 0 = None, or integer from 1 to 31 for colleges ordered alphabetically)

## Endpoints

- `/` - Redirects to Discord invite
- `/role` - Start verification process
- `/discord/callback` - Discord OAuth callback
- `/ucam/callback` - Cambridge OAuth callback

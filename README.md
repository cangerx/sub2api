<div align="center">

<img src="frontend/public/logo-transparent.png" alt="CCAPI" width="96" />

# CCAPI

[![Go](https://img.shields.io/badge/Go-1.25.7-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)

**Commercial AI Gateway Platform**

Built for teams shipping stable API products on top of multiple AI providers.

English | [中文](README_CN.md) | [日本語](README_JA.md)

</div>

## Important Notice

Please read the following carefully before using this project:

- **Terms of service risk**: Using this project may violate the terms of service of Anthropic and other upstream providers. Review each provider's agreement before use.
- **Compliant use**: Use this project only in compliance with the laws and regulations of your country or region.
- **Disclaimer**: This project is provided for technical learning and research purposes. The authors assume no liability for account bans, service interruptions, data loss, or other damages.

## Overview

CCAPI is a commercial-grade AI gateway for teams that need a stable, OpenAI-compatible control plane across multiple accounts, subscriptions, and upstream providers.

It centralizes account pooling, key distribution, model routing, billing, payments, quota visibility, media tracking, and operational monitoring. Rather than exposing upstream credentials directly, CCAPI issues platform API keys and handles authentication, scheduling, failover, usage recording, and request forwarding in one place.

## Features

| Area | Capability |
|------|------------|
| Gateway | OpenAI-compatible routing, streaming, sticky sessions, failover, model-level policies |
| Accounts | OAuth and API-key account pools for multiple upstream platforms |
| Models | Text, coding, image, video, Grok/xAI, Antigravity, Claude, Gemini, OpenAI-compatible channels |
| Billing | Token-level usage records, cost calculation, balance control, overdraw protection |
| Payments | Built-in EasyPay, Alipay, WeChat Pay, Stripe, subscription orders, recharge conversion |
| Operations | Admin dashboard, monitoring cards, usage details, media preview, account health checks |
| Security | Rate limits, concurrency limits, URL allowlists, response header filtering, compliance gate |
| Extensibility | External admin iframe integration, configurable groups, model aliases, custom tools |

## Ecosystem

Community projects that extend or integrate with CCAPI:

| Project | Description | Features |
|---------|-------------|----------|
| ~~[Sub2ApiPay](https://github.com/touwaeriol/ccapipay)~~ | ~~Self-service payment system~~ | **Now Built-in** — Payment is now integrated into CCAPI, no separate deployment needed. See [Payment Configuration Guide](docs/PAYMENT.md) |
| [ccapi-mobile](https://github.com/ckken/ccapi-mobile) | Mobile admin console | Cross-platform app (iOS/Android/Web) for user management, account management, monitoring dashboard, and multi-backend switching; built with Expo + React Native |

## Tech Stack

| Component | Technology |
|-----------|------------|
| Backend | Go 1.25.7, Gin, Ent |
| Frontend | Vue 3.4+, Vite 5+, TailwindCSS |
| Database | PostgreSQL 15+ |
| Cache/Queue | Redis 7+ |

---

## Nginx Reverse Proxy Note

When using Nginx as a reverse proxy for CCAPI (or CRS) with Codex CLI, add the following to the `http` block in your Nginx configuration:

```nginx
underscores_in_headers on;
```

Nginx drops headers containing underscores by default (e.g. `session_id`), which breaks sticky session routing in multi-account setups.

---

## Install & Deploy

Pick the path that matches your environment.

### Method 1: Docker Compose

#### Prerequisites

- Docker 20.10+
- Docker Compose v2+

#### Quick Start

```bash
# Create deployment directory
mkdir -p ccapi-deploy && cd ccapi-deploy

# Download and run deployment preparation script
curl -sSL https://raw.githubusercontent.com/cangerx/sub2api/main/deploy/docker-deploy.sh | bash

# Start services
docker compose up -d

# View logs
docker compose logs -f ccapi
```

The script downloads the production Compose template, generates secure secrets, creates local data directories, and prepares a `.env` file.

Default image:

```bash
ghcr.io/cangerx/ccapi:latest
```

Version-pinned image:

```bash
ghcr.io/cangerx/ccapi:0.1.142
```

**What the deployment script does:**
- Downloads `docker-compose.local.yml` (saved as `docker-compose.yml`) and `.env.example`
- Generates secure credentials (JWT_SECRET, TOTP_ENCRYPTION_KEY, POSTGRES_PASSWORD)
- Creates `.env` file with auto-generated secrets
- Creates data directories (uses local directories for easy backup/migration)
- Displays generated credentials for your reference

#### Manual Deployment

If you prefer manual setup:

```bash
# 1. Clone the repository
git clone https://github.com/cangerx/sub2api.git
cd sub2api/deploy

# 2. Copy environment configuration
cp .env.example .env

# 3. Edit configuration (generate secure passwords)
nano .env
```

**Required configuration in `.env`:**

```bash
# PostgreSQL password (REQUIRED)
POSTGRES_PASSWORD=your_secure_password_here

# JWT Secret (RECOMMENDED - keeps users logged in after restart)
JWT_SECRET=your_jwt_secret_here

# TOTP Encryption Key (RECOMMENDED - preserves 2FA after restart)
TOTP_ENCRYPTION_KEY=your_totp_key_here

# Optional: Admin account
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=your_admin_password

# Optional: Custom port
SERVER_PORT=8080
```

**Generate secure secrets:**
```bash
# Generate JWT_SECRET
openssl rand -hex 32

# Generate TOTP_ENCRYPTION_KEY
openssl rand -hex 32

# Generate POSTGRES_PASSWORD
openssl rand -hex 32
```

```bash
# 4. Create data directories (for local version)
mkdir -p data postgres_data redis_data

# 5. Start all services
# Option A: Local directory version (recommended - easy migration)
docker compose -f docker-compose.local.yml up -d

# Option B: Named volumes version (simple setup)
docker compose up -d

# 6. Check status
docker compose -f docker-compose.local.yml ps

# 7. View logs
docker compose -f docker-compose.local.yml logs -f ccapi
```

#### Deployment Versions

| Version | Data Storage | Migration | Best For |
|---------|-------------|-----------|----------|
| **docker-compose.local.yml** | Local directories | ✅ Easy (tar entire directory) | Production, frequent backups |
| **docker-compose.yml** | Named volumes | ⚠️ Requires docker commands | Simple setup |

**Recommendation:** Use `docker-compose.local.yml` (deployed by script) for easier data management.

#### Access

Open `http://YOUR_SERVER_IP:8080` in your browser.

If admin password was auto-generated, find it in logs:
```bash
docker compose -f docker-compose.local.yml logs ccapi | grep "admin password"
```

#### Upgrade

```bash
# Pull latest image and recreate container
docker compose -f docker-compose.local.yml pull
docker compose -f docker-compose.local.yml up -d
```

#### Easy Migration (Local Directory Version)

When using `docker-compose.local.yml`, migrate to a new server easily:

```bash
# On source server
docker compose -f docker-compose.local.yml down
cd ..
tar czf ccapi-complete.tar.gz ccapi-deploy/

# Transfer to new server
scp ccapi-complete.tar.gz user@new-server:/path/

# On new server
tar xzf ccapi-complete.tar.gz
cd ccapi-deploy/
docker compose -f docker-compose.local.yml up -d
```

#### Useful Commands

```bash
# Stop all services
docker compose -f docker-compose.local.yml down

# Restart
docker compose -f docker-compose.local.yml restart

# View all logs
docker compose -f docker-compose.local.yml logs -f

# Remove all data (caution!)
docker compose -f docker-compose.local.yml down
rm -rf data/ postgres_data/ redis_data/
```

---

### Method 2: One-Line Binary Installation

Use this mode when PostgreSQL and Redis are already available on the host or in your private network.

```bash
curl -sSL https://raw.githubusercontent.com/cangerx/sub2api/main/deploy/install.sh | sudo bash
```

The installer downloads the latest GitHub Release, installs CCAPI to `/opt/ccapi`, creates a `ccapi` systemd service, and opens the first-run setup wizard on port `8080`.

Useful commands:

```bash
sudo systemctl status ccapi
sudo journalctl -u ccapi -f
sudo systemctl restart ccapi
```

Uninstall:

```bash
curl -sSL https://raw.githubusercontent.com/cangerx/sub2api/main/deploy/install.sh | sudo bash -s -- uninstall -y
```

---

### Method 3: Pull The Image Directly

Use this if you already maintain your own PostgreSQL and Redis services.

```bash
docker pull ghcr.io/cangerx/ccapi:latest
docker pull ghcr.io/cangerx/ccapi:0.1.142
```

Supported platforms:

| Image | Architectures |
|-------|---------------|
| `ghcr.io/cangerx/ccapi:latest` | `linux/amd64`, `linux/arm64` |
| `ghcr.io/cangerx/ccapi:0.1.142` | `linux/amd64`, `linux/arm64` |

---

### Method 4: Build from Source

Build and run from source code for development or customization.

#### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+

#### Build Steps

```bash
# 1. Clone the repository
git clone https://github.com/cangerx/sub2api.git
cd sub2api

# 2. Install pnpm (if not already installed)
npm install -g pnpm

# 3. Build frontend
cd frontend
pnpm install
pnpm run build
# Output will be in ../backend/internal/web/dist/

# 4. Build backend with embedded frontend
cd ../backend
go build -tags embed -o ccapi ./cmd/server

# 5. Create configuration file
cp ../deploy/config.example.yaml ./config.yaml

# 6. Edit configuration
nano config.yaml
```

> **Note:** The `-tags embed` flag embeds the frontend into the binary. Without this flag, the binary will not serve the frontend UI.

**Key configuration in `config.yaml`:**

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "your_password"
  dbname: "ccapi"

redis:
  host: "localhost"
  port: 6379
  password: ""

jwt:
  secret: "change-this-to-a-secure-random-string"
  expire_hour: 24

default:
  user_concurrency: 5
  user_balance: 0
  api_key_prefix: "sk-"
  rate_multiplier: 1.0
```

### Sora Status (Temporarily Unavailable)

> ⚠️ Sora-related features are temporarily unavailable due to technical issues in upstream integration and media delivery.
> Please do not rely on Sora in production at this time.
> Existing `gateway.sora_*` configuration keys are reserved and may not take effect until these issues are resolved.

Additional security-related options are available in `config.yaml`:

- `cors.allowed_origins` for CORS allowlist
- `security.url_allowlist` for upstream/pricing/CRS host allowlists
- `security.url_allowlist.enabled` to disable URL validation (use with caution)
- `security.url_allowlist.allow_insecure_http` to allow HTTP URLs when validation is disabled
- `security.url_allowlist.allow_private_hosts` to allow private/local IP addresses
- `security.response_headers.enabled` to enable configurable response header filtering (disabled uses default allowlist)
- `security.csp` to control Content-Security-Policy headers
- `billing.circuit_breaker` to fail closed on billing errors
- `server.trusted_proxies` to enable X-Forwarded-For parsing
- `turnstile.required` to require Turnstile in release mode

**⚠️ Security Warning: HTTP URL Configuration**

When `security.url_allowlist.enabled=false`, the system performs minimal URL validation by default, **rejecting HTTP URLs** and only allowing HTTPS. To allow HTTP URLs (e.g., for development or internal testing), you must explicitly set:

```yaml
security:
  url_allowlist:
    enabled: false                # Disable allowlist checks
    allow_insecure_http: true     # Allow HTTP URLs (⚠️ INSECURE)
```

**Or via environment variable:**

```bash
SECURITY_URL_ALLOWLIST_ENABLED=false
SECURITY_URL_ALLOWLIST_ALLOW_INSECURE_HTTP=true
```

**Risks of allowing HTTP:**
- API keys and data transmitted in **plaintext** (vulnerable to interception)
- Susceptible to **man-in-the-middle (MITM) attacks**
- **NOT suitable for production** environments

**When to use HTTP:**
- ✅ Development/testing with local servers (http://localhost)
- ✅ Internal networks with trusted endpoints
- ✅ Testing account connectivity before obtaining HTTPS
- ❌ Production environments (use HTTPS only)

**Example error without this setting:**
```
Invalid base URL: invalid url scheme: http
```

If you disable URL validation or response header filtering, harden your network layer:
- Enforce an egress allowlist for upstream domains/IPs
- Block private/loopback/link-local ranges
- Enforce TLS-only outbound traffic
- Strip sensitive upstream response headers at the proxy

```bash
# 6. Run the application
./ccapi
```

#### Development Mode

```bash
# Backend (with hot reload)
cd backend
go run ./cmd/server

# Frontend (with hot reload)
cd frontend
pnpm run dev
```

#### Code Generation

When editing `backend/ent/schema`, regenerate Ent + Wire:

```bash
cd backend
go generate ./ent
go generate ./cmd/server
```

---

## Simple Mode

Simple Mode is designed for individual developers or internal teams who want quick access without full SaaS features.

- Enable: Set environment variable `RUN_MODE=simple`
- Difference: Hides SaaS-related features and skips billing process
- Security note: In production, you must also set `SIMPLE_MODE_CONFIRM=true` to allow startup

---

## Grok / xAI OAuth Support

CCAPI supports Grok subscription accounts through xAI OAuth and forwards OpenAI-compatible Responses traffic to xAI.

### Supported Scope

- Platform name: `grok`
- Account type: OAuth subscription accounts
- Public gateway target: `/v1/responses` and `/responses`, forwarded to `${XAI_BASE_URL:-https://api.x.ai/v1}/responses`
- Initial models: `grok-4.3`, `grok-build-0.1`, `grok-4.20-0309-reasoning`, `grok-4.20-0309-non-reasoning`, and `grok-4.20-multi-agent-0309`
- Out of scope for this provider: public Grok Chat Completions routes, image, video, TTS, transcription, browser automation, cookies, and Grok web scraping

### OAuth Configuration

The Grok OAuth flow uses PKCE and does not require committing private secrets. The default client details follow the public xAI OAuth flow used by compatible clients, and every value can be overridden by environment variable:

| Variable | Default |
|----------|---------|
| `XAI_OAUTH_CLIENT_ID` | Public xAI OAuth client ID |
| `XAI_OAUTH_SCOPE` | `openid profile email offline_access grok-cli:access api:access` |
| `XAI_OAUTH_REDIRECT_URI` | `http://127.0.0.1:56121/callback` |
| `XAI_OAUTH_AUTHORIZE_URL` | `https://auth.x.ai/oauth2/authorize` |
| `XAI_OAUTH_TOKEN_URL` | `https://auth.x.ai/oauth2/token` |
| `XAI_BASE_URL` | `https://api.x.ai/v1` |

Administrators can create or reauthorize Grok accounts from the dashboard, or use the admin API:

| Endpoint | Purpose |
|----------|---------|
| `POST /api/v1/admin/grok/oauth/auth-url` | Generate an xAI OAuth authorization URL |
| `POST /api/v1/admin/grok/oauth/exchange-code` | Exchange a callback URL, query string, or code for OAuth credentials |
| `POST /api/v1/admin/grok/oauth/refresh-token` | Validate or refresh a Grok refresh token |
| `POST /api/v1/admin/grok/accounts/:id/refresh` | Refresh an existing Grok account |

Credential storage reuses the existing account JSON fields: `access_token`, `refresh_token`, `token_type`, `expires_at`, optional `email`, optional `subscription_tier`, and `entitlement_status`.

### Usage And Quota Display

xAI quota is passive. CCAPI does not invent subscription quota values; it records whitelisted xAI rate-limit headers from successful or rate-limited upstream responses when xAI sends them. Before the first usable upstream response, the dashboard shows quota as unknown and still displays local CCAPI usage stats.

`401` responses mark the account as needing reauthorization. `403` responses are treated as entitlement or subscription-tier failures instead of token-refresh loops. `429` responses use `Retry-After` or a short cooldown to temporarily remove the account from scheduling.

---

## Antigravity Support

CCAPI supports [Antigravity](https://antigravity.so/) accounts. After authorization, dedicated endpoints are available for Claude and Gemini models.

### Dedicated Endpoints

| Endpoint | Model |
|----------|-------|
| `/antigravity/v1/messages` | Claude models |
| `/antigravity/v1beta/` | Gemini models |

### Claude Code Configuration

```bash
export ANTHROPIC_BASE_URL="http://localhost:8080/antigravity"
export ANTHROPIC_AUTH_TOKEN="sk-xxx"
```

### Hybrid Scheduling Mode

Antigravity accounts support optional **hybrid scheduling**. When enabled, the general endpoints `/v1/messages` and `/v1beta/` will also route requests to Antigravity accounts.

> **⚠️ Warning**: Anthropic Claude and Antigravity Claude **cannot be mixed within the same conversation context**. Use groups to isolate them properly.

### Known Issues

In Claude Code, Plan Mode cannot exit automatically. (Normally when using the native Claude API, after planning is complete, Claude Code will pop up options for users to approve or reject the plan.)

**Workaround**: Press `Shift + Tab` to manually exit Plan Mode, then type your response to approve or reject the plan.

---

## Project Structure

```
ccapi/
├── backend/                  # Go backend service
│   ├── cmd/server/           # Application entry
│   ├── internal/             # Internal modules
│   │   ├── config/           # Configuration
│   │   ├── model/            # Data models
│   │   ├── service/          # Business logic
│   │   ├── handler/          # HTTP handlers
│   │   └── gateway/          # API gateway core
│   └── resources/            # Static resources
│
├── frontend/                 # Vue 3 frontend
│   └── src/
│       ├── api/              # API calls
│       ├── stores/           # State management
│       ├── views/            # Page components
│       └── components/       # Reusable components
│
└── deploy/                   # Deployment files
    ├── docker-compose.yml    # Docker Compose configuration
    ├── .env.example          # Environment variables for Docker Compose
    ├── config.example.yaml   # Full config file for binary deployment
    └── install.sh            # One-click installation script
```

## Disclaimer

> **Please read carefully before using this project:**
>
> :rotating_light: **Terms of Service Risk**: Using this project may violate Anthropic's Terms of Service. Please read Anthropic's user agreement carefully before use. All risks arising from the use of this project are borne solely by the user.
>
> :book: **Disclaimer**: This project is for technical learning and research purposes only. The author assumes no responsibility for account suspension, service interruption, or any other losses caused by the use of this project.

---

## Star History

<a href="https://star-history.com/#cangerx/sub2api&Date">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/svg?repos=cangerx/sub2api&type=Date&theme=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/svg?repos=cangerx/sub2api&type=Date" />
   <img alt="Star History Chart" src="https://api.star-history.com/svg?repos=cangerx/sub2api&type=Date" />
 </picture>
</a>

---

## License

This project is licensed under the [GNU Lesser General Public License v3.0](LICENSE) (or later).

Copyright (c) 2026 Wesley Liddick

---

<div align="center">

**If you find this project useful, please give it a star!**

</div>

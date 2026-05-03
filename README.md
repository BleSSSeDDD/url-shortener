# URL Shortener

A simple URL shortener service written in Go.

## Features
- REST API (HTML + JSON)
- PostgreSQL for persistent storage
- Redis for caching
- Chi router for clean routing
- Docker containers with multi-network setup
- Traefik as reverse proxy

## Tech Stack
- **Go 1.25+** — core language
- **Chi** — lightweight router (grouping, middleware, URL params)
- **PostgreSQL 18** — main database
- **Redis 7+** — caching layer
- **Docker + Docker Compose** — containerization
- **Traefik** — reverse proxy, load balancing

## Browser View

![URL Shortener interface](https://github.com/user-attachments/assets/4f109b36-a331-40de-bbf8-f9e7c299933a)

## API Endpoints

### HTML (for humans)
- `GET /` — main page with form
- `POST /shorten` — create short link
- `GET /r/{code}` — redirect to original URL

### JSON API v1
- `GET /api` — service info
- `GET /api/v1` — v1 endpoints list
- `GET /api/v1/health` — health check
- `POST /api/v1/shorten` — create short link

  **Request:**
  ```json
  {
    "url": "https://example.com"
  }
  ```

  **Response:**
  ```json
  {
    "short_url": "http://localhost/r/abc123",
    "code": "abc123"
  }
  ```

## Prerequisites
- Docker & Docker Compose

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/BleSSSeDDD/url-shortener.git
   cd url-shortener
   ```

2. Create `.env` file (or copy from `.env.example`):
   ```bash
   cp .env.example .env
   ```

3. Start the services:
   ```bash
   docker-compose up -d
   ```

The service will be available at `http://localhost`.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | `postgres` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `12345678` |
| `DB_NAME` | Database name | `urls_and_codes` |
| `REDIS_HOST` | Redis host | `redis` |
| `REDIS_PORT` | Redis port | `6379` |

## Docker Networking Configuration for Traefik

When using Traefik with Docker Compose across multiple networks, explicit URL configuration in Traefik labels is **required** for reliable service discovery:

```yaml
labels:
  - "traefik.http.services.server-url-shortener.loadbalancer.server.url=http://url-shortener-server.url-shortener_proxy-app-network:8080"
```

**Reason:**  
1. **Multiple Networks**: The application connects to two separate Docker networks (`proxy-app-network` for Traefik routing and `app-db-network` for database communication)  
2. **Docker DNS Limitation**: While Traefik automatically discovers containers by name, it cannot determine which network interface to use when a container belongs to multiple networks  
3. **FQDN Requirement**: Using the explicit fully-qualified container name (`<service_name>.<network_name>`) ensures proper routing within the correct network segment  

**Why This Configuration Works:**  
- The container `url-shortener-server` participates in both `proxy-app-network` and `app-db-network`  
- Traefik runs only in `proxy-app-network`, so it needs the explicit network-scoped address to establish communication  
- Without this explicit URL, Traefik might resolve the container name but fail to route traffic correctly between networks  

**Alternative Approaches Considered:**  
- Using a single shared network for all services (simpler but less segmented)  
- Configuring network aliases (proved less reliable in this multi-network setup)  
- Additional Traefik healthchecks (helpful but don't solve the network routing issue)  

This configuration provides stable routing without requiring complex healthcheck dependencies between Traefik and the backend service.

## Architecture

```yaml
url-shortener/
├── cmd/
│   └── server/           # Application entry point
├── internal/
│   ├── config/           # Environment configuration
│   ├── database/         # Database connection initialization
│   ├── handlers/         # HTTP handlers and routing
│   ├── service/          # Business logic layer
│   └── storage/          # Data access layer (PostgreSQL + Redis)
├── migrations/           # Database schema migrations
├── static/               # Static assets (CSS, favicon)
├── templates/            # HTML templates
├── main.go               # Application bootstrap
├── docker-compose.yml    # Container orchestration
└── go.mod                # Go module dependencies
```

**The application follows a clean 4-layer architecture with strict separation of concerns:**

```yaml
┌─────────────────────────────────────────┐
│            HTTP Layer (handlers)        │
│  - Chi Router                           │
│  - HTML Templates                       │
│  - JSON API v1                          │
│  - Static File Serving                  │
└────────────────┬────────────────────────┘
                 │ depends on
┌────────────────▼────────────────────────┐
│        Service Layer (service)          │
│  - URL shortening logic                 │
│  - Code generation (6-char random)      │
│  - Duplicate handling                   │
│  - Cache-then-DB pattern for reads      │
└────────────────┬────────────────────────┘
                 │ depends on
┌────────────────▼────────────────────────┐
│       Storage Layer (storage)           │
│  - Cache interface (Redis)              │
│  - Postgres interface (SQL)             │
│  - Data access abstraction              │
└────────────────┬────────────────────────┘
                 │ depends on
┌────────────────▼────────────────────────┐
│      Database Layer (database)          │
│  - PostgreSQL connection init           │
│  - Redis connection init                │
│  - Connection health checks             │
└─────────────────────────────────────────┘
```

## Core Components

**Storage Layer**

```text
Cache interface — GetFromCache / AddToCache (Redis, 60ms timeout)

Postgres interface — GetUrlFromCode / SetNewPair (SQL)
```

**Service Layer**

```text
Code generation: 6 random chars from [a-zA-Z0-9] (62⁶ ≈ 56B combinations)

Set(url): INSERT with ON CONFLICT → returns existing code for duplicates

Get(code): Cache-aside pattern (Redis → PostgreSQL → populate cache)
```

**HTTP Layer**

```text
HTML: GET / (form), POST /shorten, GET /r/{code} (redirect)

JSON API v1: GET /api/v1/health, POST /api/v1/shorten

Health: GET /health → 200 OK
```

**Docker Network Design**

```text
Traefik (port 80)
    ↕ proxy-app-network
Go Server (port 8080)
    ↕ app-db-network
Redis + PostgreSQL
Two networks isolate DB traffic from proxy traffic (security).
```

**Data Flow**

```text
Shorten: Client → Handler → Service → PostgreSQL INSERT ... ON CONFLICT ... RETURNING code → Response

Redirect: Client → Handler → Service → Redis (hit?) → PostgreSQL (miss?) → Populate cache → 302 Redirect
```

**Database Schema**
```sql
urls_and_codes(url VARCHAR(500), code VARCHAR(6) PRIMARY KEY)
UNIQUE INDEX on url -- prevents duplicates
```

**Startup Sequence**

```text
Init Redis + PostgreSQL connections

Build storage → service → handlers

Start Chi server (port 8080)

Wait for SIGTERM or crash

Graceful shutdown (close DB)
```
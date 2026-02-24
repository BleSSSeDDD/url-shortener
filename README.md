Убрал структуру проекта. Вот финальный вариант:

```markdown
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
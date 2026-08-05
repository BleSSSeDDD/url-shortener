# URL Shortener

<img width="512" height="512" alt="favicon" src="https://github.com/user-attachments/assets/7c07c528-d4d0-409f-aac6-531064003c69" />

## Features
- REST API (HTML + JSON)
- PostgreSQL for persistent storage
- Redis for caching
- Chi router for clean routing
- Docker containers with multi-network setup
- Traefik as reverse proxy
- Prometheus + Grafana monitoring out of the box

## Tech Stack
- **Go 1.25+** — core language
- **Chi** — lightweight router (grouping, middleware, URL params)
- **PostgreSQL 18** — main database
- **Redis 7+** — caching layer
- **Docker + Docker Compose** — containerization
- **Traefik** — reverse proxy, load balancing
- **Prometheus + Grafana** — metrics collection and dashboards

## Browser View

![URL Shortener interface](https://github.com/user-attachments/assets/4f109b36-a331-40de-bbf8-f9e7c299933a)

## API Endpoints

### HTML
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

## Monitoring

- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000` — pre-provisioned `URL Shortener` dashboard (RPS, cache hit ratio, p95 latency)
- App metrics are exposed on an internal port (`:9100/metrics`), not reachable from outside the network — Prometheus scrapes it directly

See [docs/monitoring.md](docs/monitoring.md) for the full breakdown (metric types, cardinality, PromQL cheatsheet) and a self-check quiz.

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

## Deployment with Ansible

The `ansible/` directory provisions a bare Ubuntu host and brings the whole stack up
with a single command. Tested on Ubuntu 26.04 against three hosts in parallel.

### What the playbook does

| Role | Purpose |
|---|---|
| `common` | base packages (`curl`, `htop`, `ca-certificates`) |
| `docker_installation` | registers the official Docker apt repository together with its GPG key, installs Docker Engine, CLI, containerd and the Compose plugin, enables the service, adds the deploy user to the `docker` group |
| `url_shortener` | creates `/opt/url_shortener`, renders `.env` from a Jinja2 template, copies `docker-compose.yml`, `migrations/` and `monitoring/`, then starts the stack |

### Requirements

Control machine:

```bash
ansible-galaxy collection install community.docker
```

Target hosts: Ubuntu, SSH access by key, passwordless `sudo` for the deploy user.

### Inventory

`ansible/inventory.yml`:

```yaml
url_shortener:
  hosts:
    node1:
      ansible_host: 192.168.1.10
    node2:
      ansible_host: 192.168.1.11
  vars:
    ansible_user: user
```

Group names must not contain hyphens - Ansible exposes them as Jinja variables.

### Configuration

Application settings live in `ansible/roles/url_shortener/defaults/main.yml` and are
rendered into `.env` on the target host. Override them per host or per group through
`host_vars/` and `group_vars/`.

### Run

```bash
cd ansible
ansible-playbook -i inventory.yml url_shortener_playbook.yml
```

The playbook is idempotent - a second run over an unchanged host reports `changed=0`.

### Verify

```bash
curl http://<host>/health
```

### Note on secrets

Database credentials currently sit in `defaults/main.yml` in plain text, which is
acceptable only for a local lab. For anything beyond that they belong in
`ansible-vault`.

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

The `server` container is attached to two Docker networks:

- `proxy-app-network` - shared with Traefik, carries inbound HTTP traffic
- `app-db-network` - shared with Postgres and Redis

When a container belongs to more than one network, Traefik cannot infer which one to
use and picks one on its own. If it picks `app-db-network` - where Traefik itself is
not present - the entrypoint still accepts the TCP connection, but the request never
reaches the backend and simply times out (`HTTP 000`, not a 502).

Docker assigns subnets in network-creation order, so the pick can differ from host to
host. The same Compose file was observed working on one machine and hanging on another.

**Required configuration:**

```yaml
labels:
  - "traefik.docker.network=url-shortener_proxy-app-network"
  - "traefik.http.services.server-url-shortener.loadbalancer.server.port=8080"
```

- `traefik.docker.network` names the network Traefik must use to reach the container.
  This is what removes the ambiguity.
- `loadbalancer.server.port` supplies only the port - Traefik resolves the container
  address itself through the Docker provider.

**Pin the project name:**

```yaml
name: url-shortener
```

Compose prefixes generated network and volume names with the project name, which
defaults to the directory name. Without pinning it, deploying to `/opt/url_shortener`
produces `url_shortener_proxy-app-network` while a local checkout in `url-shortener/`
produces `url-shortener_proxy-app-network`, and any label referencing the network
breaks.

**Avoid:** hardcoding a fully qualified backend address such as
`loadbalancer.server.url=http://<container>.<network>:8080`. It duplicates
Compose-generated names inside the configuration and silently breaks whenever the
project name or deployment path changes.

## Architecture

```yaml
url-shortener/
├── cmd/
│   └── server/           # Application entry point
├── internal/
│   ├── config/           # Environment configuration
│   ├── database/         # Database connection initialization
│   ├── handlers/         # HTTP handlers and routing
│   ├── metrics/          # Prometheus metric definitions
│   ├── service/          # Business logic layer
│   └── storage/          # Data access layer (PostgreSQL + Redis)
├── migrations/           # Database schema migrations
├── monitoring/           # Prometheus config + Grafana provisioning
├── docs/                 # monitoring.md — deep dive + quiz
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
Go Server (port 8080, metrics on internal :9100)
    ↕ app-db-network
Redis + PostgreSQL + Prometheus (scrapes :9100) + Grafana
Two networks isolate DB traffic from proxy traffic (security).
Metrics port 9100 is not published — reachable only inside app-db-network.
```

**Data Flow**

```text
Shorten:
Client → Handler → Service → PostgreSQL INSERT ... ON CONFLICT ... RETURNING code → Response

Redirect:
Client → Handler → Service → Redis (hit?) → PostgreSQL (miss?) → Populate cache → 302 Redirect
```

**Database Schema**
```sql
urls_and_codes(url VARCHAR(500), code VARCHAR(6) PRIMARY KEY)
UNIQUE INDEX on url -- prevents duplicates
```

## CI/CD

The project uses GitHub Actions for automatic checking and deployment.

 When pushing to the `main` or `tests` branches, the following is run:

1. **Linter** - code quality check.
2. **Tests** - running unit tests with coverage.
3. **e2e** - building and running the full Docker Compose stack, running end-to-end tests against it.
4. **Push to Docker Hub** — automatic build and push of an image (only for `main`).

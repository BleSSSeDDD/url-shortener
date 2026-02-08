# URL Shortener

A simple URL shortener service written in Go.

## Features
- REST API
- PostgreSql database
- Docker containers

## Prerequisites
- Docker & Docker Compos

## Installation
```bash
git clone https://github.com/BleSSSeDDD/url-shortener.git
cd BleSSSeDDD/url-shortener
docker-compose up -d
```

## Browser view

<img width="1868" height="991" alt="image" src="https://github.com/user-attachments/assets/4f109b36-a331-40de-bbf8-f9e7c299933a" />


### Docker Networking Configuration for Traefik

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
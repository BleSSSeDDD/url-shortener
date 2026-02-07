# URL Shortener

A simple URL shortener service written in Go.

## Features
- REST API
- PostgreSql database
- Docker containers

## Prerequisites
- Docker & Docker Compos

## Problem solved: DNS routing in Docker multi-network setup

Problem: Gateway Timeout (504) when using Traefik with containers in multiple networks.
Cause: Docker DNS returns all container IPs, Traefik picks random one.
Solution: Use full container name with network suffix: 
`container-name.project-name_network-name:port`

## Installation
```bash
git clone https://github.com/BleSSSeDDD/url-shortener.git
cd BleSSSeDDD/url-shortener
docker-compose up -d
```

## Browser view

<img width="1868" height="991" alt="image" src="https://github.com/user-attachments/assets/4f109b36-a331-40de-bbf8-f9e7c299933a" />

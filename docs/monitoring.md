# Мониторинг: Prometheus + Grafana

Как в этом проекте устроен мониторинг, зачем каждая часть нужна и как повторить
то же самое на другом сервисе.

## Общая картина

```
[server :9100 /metrics]  ──scrape 15s──►  [Prometheus :9090]  ──PromQL──►  [Grafana :3000]
   инструментация кода                      хранит time-series             дашборд
```

**Pull-модель.** Приложение не шлёт метрики никуда — оно просто выставляет
текстовый эндпоинт `/metrics`. Prometheus сам периодически «соскребает» (scrape)
этот эндпоинт и складывает точки в свою базу временных рядов. Grafana потом
запрашивает данные у Prometheus на языке PromQL и рисует графики.

## 1. Метрики в коде — `internal/metrics/metrics.go`

Метрики объявлены как package-level переменные через `promauto`, который создаёт
метрику **и** сразу регистрирует её в глобальном реестре (его и отдаёт
`promhttp.Handler()`).

|                 Метрика                       |    Тип    |                 Что измеряет               |
|-----------------------------------------------|-----------|--------------------------------------------|
| `http_requests_total{method,route,code}`      | Counter   | число HTTP-запросов (Rate + Errors из RED) |
| `http_request_duration_seconds{method,route}` | Histogram | длительность запросов (Duration из RED)    |
| `http_requests_in_flight`                     | Gauge     | запросов «в полёте» прямо сейчас           |
| `urlshortener_urls_shortened_total`           | Counter   | число операций сокращения                  |
| `urlshortener_cache_requests_total{result}`   | Counter   | обращения к кэшу (hit/miss)                |

**Четыре типа метрик:**
- **Counter** — только растёт (сбрасывается на рестарте). По нему смотрят `rate()`.
- **Gauge** — растёт и падает (текущее состояние).
- **Histogram** — раскладывает наблюдения по «корзинам» (buckets); из них PromQL
  считает перцентили (p95/p99). Для латентности нужен именно он — среднее «врёт».
- **Summary** — тут не используется.

**Кардинальность (главное правило).** В метки (labels) кладём только величины с
**ограниченным** числом значений: method, шаблон маршрута, статус-код. Нельзя
класть безграничное — короткий код ссылки, user id и т.п.: каждое новое значение
= новая серия в памяти → «взрыв кардинальности».

## 2. RED-метрики через middleware — `internal/handlers/metrics_middleware.go`

Middleware `Metrics` оборачивает весь роутер (`r.Use(Metrics)`), поэтому метрики
снимаются со всех хендлеров разом. Что важно:

- **Статус-код** перехватываем через `middleware.NewWrapResponseWriter` (обычный
  `ResponseWriter` его не помнит).
- **Метка `route`** берётся из `chi.RouteContext(r.Context()).RoutePattern()`
  **после** `next.ServeHTTP` — это шаблон `/r/{code}`, а не конкретный URL. Так все
  запросы к разным коротким ссылкам группируются в одну серию (см. кардинальность).
- **In-flight** крутится через `Inc()` / `defer Dec()` — defer гарантирует `-1`
  даже при панике.

Бизнес-метрики (`internal/service/shortener.go`): в `Get` считаем `hit`/`miss`,
в `Set` — `urls_shortened_total`.

## 3. `/metrics` на отдельном порту 9100

Публичный сервер слушает `:8080` (через него ходят пользователи и traefik).
Метрики отдаёт **второй** `http.Server` на `:9100`, запущенный в отдельной
горутине (см. `Start()`). Порт 9100 **не публикуется** наружу — Prometheus ходит
к нему по внутренней Docker-сети. Это стандартный прод-подход: `/metrics` в
принципе недостижим извне, и не нужно прятать его на уровне reverse proxy.

> Раньше `/metrics` висел на 8080 и торчал наружу через traefik; закрывали его
> middleware `ipAllowList`. С отдельным портом этот костыль не нужен.

## 4. Prometheus — `monitoring/prometheus.yml`

```yaml
global:
  scrape_interval: 15s
scrape_configs:
  - job_name: url-shortener
    static_configs:
      - targets: ['server:9100']
```

`server:9100` — имя сервиса из docker-compose (Docker резолвит его в IP по
внутренней сети). Путь `/metrics` и схему `http` Prometheus подставляет сам.

## 5. Grafana — provisioning (инфраструктура как код)

Grafana поднимается уже настроенной, без ручных кликов, из папки
`monitoring/grafana/provisioning`:
- `datasources/datasource.yml` — источник Prometheus. У него **запинен**
  `uid: prometheus`: без постоянного тома Grafana иначе генерит источнику новый
  случайный id при каждом рестарте, и дашборд с хардкодным id «слепнет».
- `dashboards/dashboards.yml` — провайдер, который грузит все `*.json` из папки.
- `dashboards/url-shortener.json` — сам дашборд (экспорт из UI). Все панели в нём
  ссылаются на `uid: prometheus`.

Анонимный доступ с ролью Admin включён через переменные окружения — **только для
удобства учёбы**, в проде так нельзя.

## Как запустить

```bash
cp .env.example .env
docker compose -f docker-compose.yml -f docker-compose.ci.yml up -d --build
```

- Приложение: `http://localhost` (через traefik)
- Prometheus: `http://localhost:9090` (Status → Targets → `server` должен быть UP)
- Grafana: `http://localhost:3000` → дашборд **URL Shortener**

Нагнать трафик для наглядности: см. `scripts` или просто подёргать
`POST /api/v1/shorten` и `GET /r/{code}`.

## Шпаргалка PromQL

| Запрос | Что показывает |
|--------|----------------|
| `sum(rate(http_requests_total[5m]))` | суммарный RPS |
| `sum by (route) (rate(http_requests_total[5m]))` | RPS по маршрутам |
| `sum(rate(http_requests_total{code=~"5.."}[5m]))` | скорость 5xx-ошибок |
| `sum(rate(urlshortener_cache_requests_total{result="hit"}[5m])) / sum(rate(urlshortener_cache_requests_total[5m]))` | cache hit ratio |
| `histogram_quantile(0.95, sum by (le) (rate(http_request_duration_seconds_bucket[5m])))` | p95 задержки |

`rate(counter[5m])` = средняя скорость роста счётчика за окно; по счётчикам почти
всегда берут `rate`, а не сырое значение.

## Тесты

- `internal/handlers/metrics_middleware_test.go` — middleware считает по шаблону
  маршрута; `/metrics` отдаёт метрики.
- `internal/service/metrics_test.go` — `Get` инкрементит `hit`/`miss`.

Оба используют `prometheus/client_golang/prometheus/testutil`.

---
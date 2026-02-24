# =============================================
# URL-SHORTENER TEST SCRIPT
# =============================================

Write-Host "================================`n" -ForegroundColor Green
Write-Host "STARTING TESTS" -ForegroundColor Green
Write-Host "================================`n" -ForegroundColor Green

# 1. Health checks
Write-Host "1. Health checks:" -ForegroundColor Cyan
Write-Host "----------------------------------------"

try {
    $health = Invoke-RestMethod "http://localhost/health" -ErrorAction Stop
    Write-Host "   OK /health: $health" -ForegroundColor Green
} catch {
    Write-Host "   FAIL /health: $_" -ForegroundColor Red
}

try {
    $healthApi = Invoke-RestMethod "http://localhost/api/v1/health" -ErrorAction Stop
    Write-Host "   OK /api/v1/health: $($healthApi | ConvertTo-Json -Compress)" -ForegroundColor Green
} catch {
    Write-Host "   FAIL /api/v1/health: $_" -ForegroundColor Red
}

Write-Host ""

# 2. API info endpoints
Write-Host "2. API info endpoints:" -ForegroundColor Cyan
Write-Host "----------------------------------------"

try {
    $api = Invoke-RestMethod "http://localhost/api" -ErrorAction Stop
    Write-Host "   OK /api:" -ForegroundColor Green
    Write-Host "      Service: $($api.service)" -ForegroundColor Gray
    Write-Host "      Versions: $($api.versions -join ', ')" -ForegroundColor Gray
} catch {
    Write-Host "   FAIL /api: $_" -ForegroundColor Red
}

try {
    $apiV1 = Invoke-RestMethod "http://localhost/api/v1" -ErrorAction Stop
    Write-Host "   OK /api/v1:" -ForegroundColor Green
    Write-Host "      Version: $($apiV1.version)" -ForegroundColor Gray
    Write-Host "      Status: $($apiV1.status)" -ForegroundColor Gray
} catch {
    Write-Host "   FAIL /api/v1: $_" -ForegroundColor Red
}

Write-Host ""

# 3. Create short links
Write-Host "3. Create short links:" -ForegroundColor Cyan
Write-Host "----------------------------------------"

$testUrls = @(
    "https://google.com",
    "https://github.com",
    "https://stackoverflow.com"
)

$codes = @{}

foreach ($url in $testUrls) {
    try {
        $body = @{url = $url} | ConvertTo-Json
        $result = Invoke-RestMethod -Uri "http://localhost/api/v1/shorten" -Method Post -Body $body -ContentType "application/json" -ErrorAction Stop
        $codes[$url] = $result.code
        Write-Host "   OK $url -> $($result.short_url)" -ForegroundColor Green
    } catch {
        Write-Host "   FAIL $url : $_" -ForegroundColor Red
    }
}

Write-Host ""

# 4. Uniqueness test (should return same code)
Write-Host "4. Uniqueness test:" -ForegroundColor Cyan
Write-Host "----------------------------------------"

$firstUrl = $testUrls[0]
try {
    $body = @{url = $firstUrl} | ConvertTo-Json
    $result2 = Invoke-RestMethod -Uri "http://localhost/api/v1/shorten" -Method Post -Body $body -ContentType "application/json" -ErrorAction Stop
    if ($result2.code -eq $codes[$firstUrl]) {
        Write-Host "   OK $firstUrl -> code $($result2.code) (same as before)" -ForegroundColor Green
    } else {
        Write-Host "   FAIL Code changed: was $($codes[$firstUrl]), now $($result2.code)" -ForegroundColor Red
    }
} catch {
    Write-Host "   FAIL $_" -ForegroundColor Red
}

Write-Host ""

# 5. Redirect test
Write-Host "5. Redirect test:" -ForegroundColor Cyan
Write-Host "----------------------------------------"

$testCode = $codes[$firstUrl]
if ($testCode) {
    try {
        $redirect = Invoke-WebRequest -Uri "http://localhost/r/$testCode" -MaximumRedirection 0 -ErrorAction SilentlyContinue
        Write-Host "   FAIL Expected redirect, got $($redirect.StatusCode)" -ForegroundColor Red
    } catch {
        if ($_.Exception.Response.StatusCode -eq 302) {
            $location = $_.Exception.Response.Headers.Location
            Write-Host "   OK /r/$testCode -> $location" -ForegroundColor Green
        } else {
            Write-Host "   FAIL $_" -ForegroundColor Red
        }
    }
} else {
    Write-Host "   SKIP No code to test" -ForegroundColor Yellow
}

Write-Host ""

# 6. Static files
Write-Host "6. Static files:" -ForegroundColor Cyan
Write-Host "----------------------------------------"

try {
    $style = Invoke-WebRequest "http://localhost/static/style.css" -ErrorAction Stop
    if ($style.Content -match "body") {
        Write-Host "   OK /static/style.css loaded" -ForegroundColor Green
    } else {
        Write-Host "   FAIL /static/style.css corrupted" -ForegroundColor Red
    }
} catch {
    Write-Host "   FAIL /static/style.css: $_" -ForegroundColor Red
}

Write-Host ""

# 7. Main page
Write-Host "7. Main page:" -ForegroundColor Cyan
Write-Host "----------------------------------------"

try {
    $homePage = Invoke-WebRequest "http://localhost/" -ErrorAction Stop
    if ($homePage.Content -match "Сокращатель ссылок") {
        Write-Host "   OK Main page loaded" -ForegroundColor Green
    } else {
        Write-Host "   FAIL Main page corrupted" -ForegroundColor Red
    }
} catch {
    Write-Host "   FAIL Main page: $_" -ForegroundColor Red
}

Write-Host ""

# 8. Docker containers
Write-Host "8. Docker containers:" -ForegroundColor Cyan
Write-Host "----------------------------------------"

$containers = @("url-shortener-server", "url-shortener-redis", "url-shortener-postgres", "url-shortener-traefik")
foreach ($container in $containers) {
    $status = docker ps --filter "name=$container" --format "table {{.Status}}" | Select-Object -Skip 1
    if ($status -match "Up|Healthy") {
        Write-Host "   OK $container — $status" -ForegroundColor Green
    } else {
        Write-Host "   FAIL $container — not running" -ForegroundColor Red
    }
}

Write-Host ""

# 9. Redis cache (optional)
Write-Host "9. Redis cache (optional):" -ForegroundColor Cyan
Write-Host "----------------------------------------"

try {
    $redisTest = docker exec url-shortener-redis redis-cli ping 2>$null
    if ($redisTest -eq "PONG") {
        Write-Host "   OK Redis responding" -ForegroundColor Green
        
        if ($testCode) {
            $cached = docker exec url-shortener-redis redis-cli GET $testCode 2>$null
            if ($cached) {
                Write-Host "   OK Code $testCode cached: $cached" -ForegroundColor Green
            } else {
                Write-Host "   WARN Code $testCode not in cache" -ForegroundColor Yellow
            }
        }
    } else {
        Write-Host "   WARN Redis not accessible" -ForegroundColor Yellow
    }
} catch {
    Write-Host "   WARN Could not check Redis" -ForegroundColor Yellow
}

Write-Host ""

# 10. Summary
Write-Host "================================`n" -ForegroundColor Green
Write-Host "TESTS COMPLETED" -ForegroundColor Green
Write-Host "Check output above for errors" -ForegroundColor Yellow
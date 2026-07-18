// Package e2e содержит end-to-end тесты, которые гоняются против уже поднятого
// стека (docker compose): реальные HTTP-запросы к сервису, БД и кэшу.
//
// Сами тесты живут под build-тегом e2e, поэтому в обычный `go test ./...` не
// попадают (нужен живой сервер). Запуск:
//
//	BASE_URL=http://localhost:80 go test -tags e2e -v ./test/e2e/...
package e2e

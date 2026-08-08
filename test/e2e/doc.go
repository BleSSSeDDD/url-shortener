// Package e2e holds end-to-end tests that run against an already running
// stack (docker compose): real HTTP requests hitting the service, database and cache.
//
// The tests sit behind the e2e build tag, so a plain `go test ./...` does not
// pick them up (a live server is required). Run with:
//
//	BASE_URL=http://localhost:80 go test -tags e2e -v ./test/e2e/...
package e2e

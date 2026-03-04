.PHONY: build build-web build-server clean dev dev-web dev-server

# Production single binary (with embedded web)
build: build-web build-server

build-web:
	cd web && npm ci && npm run build
	rm -rf internal/api/web_dist
	cp -r web/build internal/api/web_dist

build-server: build-web
	go build -tags production -o go-jira-server ./cmd/go-jira-server

# CLI only (no web)
build-cli:
	go build -o go-jira ./cmd/go-jira

# Development
dev-server:
	go run ./cmd/go-jira-server --port 8080 --cors-origin http://localhost:5173

dev-web:
	cd web && npm run dev

# Clean
clean:
	rm -f go-jira go-jira-server
	rm -rf internal/api/web_dist
	rm -rf web/build web/node_modules

.PHONY: build
build:
	go build -o bin/fs.tool cmd/fs/main.go

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run ./...

.PHONY: format
format:
	go fmt ./...
	go mod tidy

.PHONY: clean
clean:
	rm -f bin/fs.tool
	go clean ./...

.PHONY: deps
deps:
	go mod tidy

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: bump-tools-yaml-version
bump-tools-yaml-version:
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "Error: Cannot bump version to 'dev'. Set VERSION explicitly: make publish VERSION=0.3.0"; \
		exit 1; \
	fi
	@echo "Updating tool.yaml version to $(VERSION)..."
	@sed -i.bak 's/^version: .*/version: $(VERSION)/' tool.yaml && rm -f tool.yaml.bak

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev")
.PHONY: publish
publish: bump-tools-yaml-version build
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "Error: Cannot publish 'dev' version. Set VERSION explicitly: make publish VERSION=0.3.0"; \
		exit 1; \
	fi
	@echo "Building binary for release..."
	@if [ ! -f bin/fs.tool ]; then \
		echo "Error: Binary build failed"; \
		exit 1; \
	fi
	@echo "Staging binary for commit..."
	@git add -f bin/fs.tool tool.yaml
	@if ! git diff --cached --quiet; then \
		echo "Committing changes for v$(VERSION)..."; \
		git commit -m "Release v$(VERSION)"; \
	fi
	@echo "Creating tag v$(VERSION)..."
	git tag v$(VERSION)
	@echo "Pushing tag v$(VERSION)..."
	git push origin v$(VERSION)
	@echo "Pushing commits..."
	git push origin main
	@echo "Release v$(VERSION) published successfully!"
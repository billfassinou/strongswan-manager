SHELL := /bin/bash
GO_IMAGE := golang:1.23
export GOFLAGS := -buildvcs=false

# Utilise le go(1) local s'il est présent, sinon retombe sur une image Docker golang.
ifeq ($(shell command -v go 2>/dev/null),)
GORUN := docker run --rm -v $(PWD):/src -w /src -e GOFLAGS=-buildvcs=false $(GO_IMAGE)
else
GORUN :=
endif

.PHONY: tidy build web test test-integration cover lint run lab-up lab-down logs fmt vet

web: ## Installe et build la SPA React (requis avant un build go local, embarquée dans le binaire)
	cd web && npm install && npm run build

tidy: ## Résout les dépendances et génère go.sum
	$(GORUN) go mod tidy

build: ## Compile le binaire
	$(GORUN) go build ./...

vet: ## Analyse statique
	$(GORUN) go vet ./...

test: ## Tests unitaires (aucune dépendance externe)
	$(GORUN) go test ./...

cover: ## Tests unitaires + rapport de couverture
	$(GORUN) go test -cover ./...

test-integration: ## Tests d'intégration (dossier ./test, boîte noire) contre un postgres jetable
	@bash -c 'set -e; \
	  docker rm -f swan-test-pg >/dev/null 2>&1 || true; \
	  docker run -d --name swan-test-pg -e POSTGRES_USER=swan -e POSTGRES_PASSWORD=swan -e POSTGRES_DB=swan -p 55432:5432 postgres:16-alpine >/dev/null; \
	  echo "attente de postgres..."; until docker exec swan-test-pg pg_isready -U swan >/dev/null 2>&1; do sleep 1; done; \
	  trap "docker rm -f swan-test-pg >/dev/null 2>&1 || true" EXIT; \
	  DATABASE_URL="postgres://swan:swan@localhost:55432/swan?sslmode=disable" go test -tags integration -count=1 ./test/'

fmt: ## Formatte le code
	$(GORUN) gofmt -w .

run: ## Démarre postgres + backend (adaptateur VICI mock)
	docker compose up --build

lab-up: ## Démarre le lab complet avec 2 passerelles strongSwan (VICI réel)
	VICI_ENDPOINTS="gw-a=unix:/gw/a/charon.vici,gw-b=unix:/gw/b/charon.vici" \
	CRL_URL="http://backend:8080/crl.der" CRL_VALIDITY="30s" \
		docker compose --profile lab up --build -d

lab-down: ## Arrête et nettoie le lab
	docker compose --profile lab down -v

logs: ## Suit les logs du backend
	docker compose logs -f backend

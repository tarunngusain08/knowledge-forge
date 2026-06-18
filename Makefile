GO ?= go
COMPOSE ?= docker compose

.PHONY: test vet tidy run-api run-worker migrate-up compose-up compose-down sqlc goose-status validate-acceptance

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

run-api:
	$(GO) run ./cmd/api

run-worker:
	$(GO) run ./cmd/worker

migrate-up:
	$(GO) run ./cmd/migrate up

compose-up:
	$(COMPOSE) up --build

compose-down:
	$(COMPOSE) down

sqlc:
	$(GO) run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

goose-status:
	$(GO) run github.com/pressly/goose/v3/cmd/goose@v3.24.3 -dir migrations postgres "$$DATABASE_URL" status

validate-acceptance:
	python3 eval-runner/acceptance/validation_runner.py \
		--fixtures eval-runner/acceptance/fixtures/acceptance-suite.json \
		--candidate eval-runner/acceptance/candidates/passing-candidate.json \
		--output eval-runner/acceptance/reports
	python3 -m unittest discover eval-runner/acceptance -p 'test_*.py'

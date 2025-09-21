# Repository Guidelines

## Project Structure
- `backend/`: Echo-based public API. Domain rules live in `internal/domain`, HTTP transport in `internal/delivery/http`, and deployment assets under `configs/` and `bin/`.
- `local-backend/`: Always-on worker that orchestrates RabbitMQ, Claude, and SQLite. Tests and benchmarks sit in `tests/`, automation in `scripts/`.
- `frontend/`: Next.js UI. Core code is in `src/`, Playwright suites in `tests/`, shared UI primitives in `src/components`, and static assets in `public/`.
- Reference `docs/` and the top-level playbooks (`DEPLOYMENT.md`, `PERFORMANCE_REPORT.md`) when changing system behaviour.

## Build, Test & Development Commands
- API server: `go run cmd/server/main.go` for local dev, `go build -o bin/server cmd/server/main.go` for releases, `go test ./...` before commits.
- Worker: `go run main.go` starts the queue processor; `./scripts/run-tests.sh` runs fmt/vet, unit, integration, e2e, and benchmarks with ≥90% coverage enforced.
- Frontend: `npm run dev` for live reload, `npm run build && npm run start` for production previews, `npm run lint`, `npm run test`, and `npm run test:e2e` to mirror CI.

## Coding Style & Naming
- Go code must stay `gofmt` clean (tabs, grouped imports) and pass `go vet`; keep package names nouns (`session`, `config`) and interface contracts in `internal/domain/repositories`.
- TypeScript follows ESLint’s `next/core-web-vitals`. Use PascalCase for React components, camelCase for hooks/utilities, and kebab-case route folders under `src/app`.
- Tailwind is the default styling system; prefer utility-first classes and record shared tokens in `components.json`.

## Testing Expectations
- Co-locate Go tests as `*_test.go`; tag integration specs with `//go:build integration` and place them in `local-backend/tests/integration`. Retain generated coverage artefacts under `coverage/`.
- Frontend unit tests live in `frontend/tests/unit` using Jest and Testing Library; update snapshots intentionally and note diffs in reviews. Execute `npm run test:coverage` for merge-ready PRs.
- Playwright specs in `frontend/tests/e2e` require an active API; capture trace bundles when chasing regressions.

## Commit & PR Workflow
- Follow the existing history: imperative subjects, optionally scoped with the issue reference (`Issue #66: Complete application services layer implementation`).
- PRs should link issues, describe behavioural changes, attach test output, and include screenshots or recordings for UI work. Flag migrations or config updates in bold.

## Configuration & Security
- Never commit secrets; seed `.env` files from the provided templates in `backend/` and `local-backend/`. Use environment variables for Claude credentials and database DSNs.
- Health endpoints (e.g., `/api/v1/health`) expose runtime metrics—redact before sharing logs externally.

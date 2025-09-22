# Repository Guidelines

## Project Structure & Module Organization
- `backend/`: Echo API with domain logic in `internal/domain`, HTTP adapters in `internal/delivery/http`, and deploy assets in `configs/` and `bin/`.
- `local-backend/`: Queue worker coordinating RabbitMQ, Claude, and SQLite; automation lives under `scripts/`, and benches/tests under `tests/`.
- `frontend/`: Next.js app with feature code in `src/`, shared primitives in `src/components`, Playwright suites in `tests/`, and static assets in `public/`.
- Cross-team references sit in `docs/`, while `DEPLOYMENT.md` and `PERFORMANCE_REPORT.md` capture runbooks and perf baselines.

## Build, Test, and Development Commands
- API: `go run cmd/server/main.go` for local dev, `go build -o bin/server cmd/server/main.go` for release binaries, `go test ./...` ahead of commits.
- Worker: `go run main.go` spins up the orchestrator; `./scripts/run-tests.sh` chains fmt/vet, unit, integration, e2e, and benchmark suites with ≥90% coverage strictness.
- Frontend: `npm run dev` for HMR, `npm run build && npm run start` for prod previews, plus `npm run lint`, `npm run test`, and `npm run test:e2e` to match CI.

## Coding Style & Naming Conventions
- Go sources must stay `gofmt` clean, pass `go vet`, and keep package names as nouns (e.g., `session`, `config`); repository interfaces belong in `internal/domain/repositories`.
- TypeScript follows ESLint `next/core-web-vitals`: PascalCase components, camelCase hooks and utilities, kebab-case route directories under `src/app`.
- Tailwind is the default styling system; record shared tokens in `components.json` and prefer utility-first class composition.

## Testing Guidelines
- Place Go tests alongside implementations as `*_test.go`; gate integration specs with `//go:build integration` and store them in `local-backend/tests/integration`.
- Frontend unit coverage runs through Jest + Testing Library in `frontend/tests/unit`; execute `npm run test:coverage` before publishing changes.
- Playwright E2E suites live in `frontend/tests/e2e`; run them against a live API and capture trace bundles when debugging.

## Commit & Pull Request Guidelines
- Write imperative commit subjects, optionally scoped with issue references (e.g., `Issue #66: Complete application services layer implementation`).
- PRs should link issues, explain behaviour shifts, attach test output, and include screenshots or recordings for UI changes; highlight migrations or config updates in **bold**.

## Security & Configuration Tips
- Never commit secrets; seed `.env` files from templates in `backend/` and `local-backend/`, and rely on env vars for Claude credentials and DSNs.
- Health endpoints such as `/api/v1/health` expose runtime metrics—sanitize logs before sharing externally.

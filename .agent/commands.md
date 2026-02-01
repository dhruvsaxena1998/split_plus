# Commands Reference

Task runner: `just`

## Backend (Go)

| Command | Description |
|---------|-------------|
| `just test` | Run all tests |
| `just test -run TestName` | Run single test |
| `just test-coverage` | Run tests with coverage |
| `just test-e2e` | Run integration tests |
| `just be-fmt` | Format Go code |
| `just be-lint` | Lint Go code |
| `just go-build` | Build binaries |
| `just sqlc-generate` | Generate sqlc code |

## Database

| Command | Description |
|---------|-------------|
| `just migrate-up` | Apply migrations |
| `just migrate-down` | Rollback migrations |
| `just migrate-status` | Check migration status |
| `just migrate-create name` | Create new migration |

## Frontend (React/TypeScript)

| Command | Description |
|---------|-------------|
| `just fe-dev` | Development server |
| `just fe-build` | Build for production |
| `just fe-test` | Run tests |
| `just fe-lint` | Lint code |
| `just fe-format` | Format code |
| `just fe-check` | Format + lint |

## Docker

| Command | Description |
|---------|-------------|
| `just up` | Start all services |
| `just down` | Stop all services |
| `just restart` | Rebuild and restart |
| `just logs` | View logs |

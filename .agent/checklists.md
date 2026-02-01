# Development Checklists

## Adding a New Backend Feature

1. Create migration: `just migrate-create name`
2. Add SQL queries to `internal/db/queries/`
3. Generate code: `just sqlc-generate`
4. Create repository interface + implementation
5. Create service interface + implementation
6. Create handler with request/response structs
7. Add routes with functional options
8. **Write tests** (mandatory)
9. Wire dependencies in `internal/app/app.go`
10. Verify: `just test`

## Before Committing

### Backend
- [ ] Tests pass: `just test`
- [ ] Code formatted: `just be-fmt`
- [ ] Code linted: `just be-lint`
- [ ] Tests written for new code

### Frontend
- [ ] Tests pass: `just fe-test`
- [ ] Code formatted: `just fe-format`
- [ ] Code linted: `just fe-lint`

### General
- [ ] No hardcoded secrets/credentials
- [ ] Error handling implemented

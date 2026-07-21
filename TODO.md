# Implementation: `requires_exit_scan` for Sensitive Zones

## Step-by-step Plan

- [x] 1. Create migration files (up/down)
- [x] 2. Update `internal/models/models.go` — add `RequiresExitScan` field
- [x] 3. Update `internal/access/errors.go` — add `ErrRequiresExitScan`
- [x] 4. Update `internal/access/interfaces.go` — add `GetRequiresExitScan` to `ZoneRepository`
- [x] 5. Update `internal/access/access_repo.go` — implement method + update queries
- [x] 6. Update `internal/access/access_service.go` — add check before auto-exit
- [x] 7. Update `internal/httpdir/access_handlers.go` — accept new field, handle new error
- [x] 8. Update `internal/httpdir/response.go` — add `RequiresExitScan` to `ZoneResponse`
- [x] 9. Update `internal/access/access_service_test.go` — add two new tests


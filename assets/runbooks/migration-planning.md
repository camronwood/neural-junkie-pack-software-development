# Migration planning runbook

**Trigger:** `@SoftwareArchitect` + `@DatabaseSpecialist` for schema or service boundary changes.

## Steps

1. **Current state** — Document existing schema/services and consumers.
2. **Target state** — Define new boundaries, tables, and API contracts.
3. **Schema diff** — Use `validate_schema` / `generate_migration` tools on proposed DDL.
4. **Rollout** — Plan expand/contract phases, backfill, and feature flags.
5. **Rollback** — Define revert path and data reconciliation checks.

# Incident handoff runbook

**Trigger:** `@IncidentManager` → engineering specialists.

## Steps

1. **Triage** — IncidentManager fetches ticket, summarizes impact and severity (P0–P4).
2. **Reproduce** — Capture stack trace, request path, and environment; link Sentry if available.
3. **Engineering** — Hand off to `@BackendEngineer` or `@SREObservabilityEngineer` for root cause.
4. **Fix proposal** — Specialist proposes patch or mitigation; CodeReviewer optional for risky changes.
5. **Jira update** — IncidentManager posts status, root cause summary, and next steps on the ticket.

## Jira comment template

```
## Update from Neural Junkie

**Status:** Investigating | Mitigated | Resolved
**Root cause:** …
**Next steps:** …
**Owner:** @…
```

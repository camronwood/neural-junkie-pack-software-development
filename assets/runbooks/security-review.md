# Security review runbook

**Trigger:** `@SecurityReviewer` or keywords: security, auth, CVE, OWASP.

## Steps

1. **Scope** — Identify changed files and entrypoints (API routes, auth middleware).
2. **Scan** — Run `run_gosec` / `run_npm_audit` / `scan_secrets` via SecurityReviewer MCP tools.
3. **Threat model** — Note trust boundaries, secrets handling, and injection surfaces.
4. **AWS IAM** — If changes touch cloud IAM or policies, consult `@AWSExpert`.
5. **Sign-off** — Post findings with severity (critical/high/medium/low) and required fixes before merge.

# Coffer

## General Overview

Coffer is a small Go service that stores information about paying patrons and ledger balances in a local SQLite database. A command line client (`coffer`) communicates with the HTTP server (`coffer-server`) to post ledger transactions, manage API keys and allocation rules, and to fetch summary metrics. The server also processes Stripe webhooks to keep patron records and payments in sync. All operations that modify state require an API key.

## Important Design Elements

- **Environment driven configuration** - Paths to the database file and listen port are taken from environment variables (`DB_FILE_PATH` and `PORT`). Additional credentials such as the Stripe key, webhook secret and bootstrap API key are loaded from files under a directory specified by `CREDENTIALS_DIRECTORY`. Optional Cross-Origin Resource Sharing origins can be configured via `CORS_ALLOWED_ORIGINS` as a comma separated list. This conforms with the way `systemd` exposes credentials to services. `systemd` unit files are provided out of the box in the `init` folder.
- **Pluggable storage via interfaces** - The `service` package exposes interfaces for the ledger, patrons, allocation rules, metrics and Stripe events. Actual persistence uses the `internal/database` package, but the design allows other storage layers.
- **SQLite schema initialization** - On startup the server opens the database and creates tables for customers, subscriptions, payments, payouts, transactions, allocation rules and API keys if they do not already exist. Default allocation rules are inserted when none are present.
- **API key management** - API tokens are salted and hashed in the database. A bootstrap key can be provided for first run. New keys are created and revoked through the `/settings/keys` endpoints.
- **Stripe integration** - Webhook payloads are validated using the Stripe signature secret. Events update the customer, subscription, payment and payout tables and post ledger entries for successful payments.
- **Allocation based ledger posting** - Each payment is split across one or more ledgers using configurable percentage rules. The rules must sum to 100 percent.
- **Authentication middleware** - Mutating endpoints require the `Authorization: Bearer` header. Tokens are verified against the stored API keys before the request is forwarded.

## HTTP API Overview

All responses use the form:

```json
{
  "error": {"code": <int>, "message": <string>},
  "data": <payload or null>
}
```

### `/health`
`GET` - Returns both the general `status` of the program (`ok`, if reachable), and a separate check on the status of the database (`ok` or `unreachable`).

### `/ledger/{ledger}`
`GET` - Returns a ledger snapshot. Optional query parameters `since` and `until` accept dates in `YYYY-MM-DD` format.

### `/ledger/{ledger}/transactions`
`GET` - List transactions with optional `limit` and `offset` query parameters.
`POST` - Create a transaction. Body fields: `date` (RFC3339), `amount`, `label`. Requires a Bearer token.

### `/metrics`
`GET` - Returns summary metrics about active patrons and revenue.

### `/patrons`
`GET` - Lists patrons. Supports `limit` and `offset` parameters. Requires a Bearer token.

### `/settings/allocations`
`GET` - Retrieve allocation rules.
`PUT` - Replace all allocation rules with an array of `{id, ledger, percentage}`. Percentages must total 100. Requires a Bearer token.

### `/settings/keys`
`POST` - Create a new API key. Returns the token once. Requires a Bearer token.
`DELETE /settings/keys/{id}` - Delete an API key. Requires a Bearer token.

### `/stripe/webhook`
`POST` - Stripe webhook endpoint. The signature header is validated and the event is queued for processing. Returns `200 OK` if accepted.

Authentication for modifying routes uses `Authorization: Bearer <token>` issued by `/settings/keys`.

## Running Tests

Run all unit tests with:

```sh
go test ./...
```

All current tests should pass.

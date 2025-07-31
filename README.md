# Coffer

## General Overview

Coffer consists of a small HTTP server (`coffer-server`) and a command line
client (`coffer`). The server exposes a JSON API that stores information about
customers, Stripe subscriptions/payments, and ledger transactions in a local
SQLite database. Incoming Stripe webhooks are processed and recorded. The API
allows authorized clients to create transactions, manage API keys and allocation
rules, and fetch metrics about active patrons. The CLI is a thin wrapper around
these endpoints for local administration.

## Important Design Elements

- **Environment driven configuration** – `coffer-server` expects several
  environment variables and credential files on start up. Paths to the database
  and listening port are read from `DB_FILE_PATH` and `PORT`, while Stripe and
  API key credentials are loaded from a directory specified by
  `CREDENTIALS_DIRECTORY`【F:cmd/coffer-server/main.go†L18-L26】.

- **Pluggable storage via interfaces** – The `service` package defines
  interfaces for key domains (ledger, patrons, allocations, metrics, Stripe
  events). Concrete implementations live in `internal/database`, keeping
  business logic separate from persistence. This design allows alternative
  storage layers if needed.

- **SQLite schema initialization** – On start the database is opened and a set of
  tables is created if they do not exist. Tables include `customer`,
  `subscription`, `payment`, `payout`, `tx` (ledger transactions), `allocation`
  and `api_key`【F:internal/database/database.go†L42-L99】. Default allocation
  rules are inserted when no rules are present【F:internal/database/allocations.go†L13-L33】.

- **API key management** – API tokens are stored salted and hashed. Tokens are
  generated via `CreateAPIKey` and verified in the request middleware. Keys are
  only initialized from a bootstrap credential if the database is empty
  (allowing manual rotation)【F:internal/service/keys.go†L24-L62】【F:internal/api/auth.go†L10-L33】.

- **Stripe integration** – Webhooks are validated using a shared secret and the
  Stripe client library. Events trigger updates to customer, subscription,
  payment or payout tables and ledger entries when payments succeed. Updates
  are queued to avoid redundant processing【F:internal/service/stripe.go†L24-L187】【F:internal/service/stripe.go†L200-L281】.

- **Allocation based ledger posting** – Payment amounts are split across one or
  more ledgers according to allocation rules. Each rule specifies a ledger name
  and percentage; the sum must equal 100% when updated. Transactions are
  recorded with labels indicating origin.

- **Authentication middleware** – Endpoints that mutate state require a Bearer
  token. The middleware extracts the token, verifies it against the stored API
  keys and rejects unauthorized requests【F:internal/api/auth.go†L10-L33】.

## HTTP API Overview

All responses share the structure below where `data` varies by endpoint and
`error` is non‑null on failures:

```json
{
  "error": {"code": <int>, "message": <string>},
  "data":  <payload or null>
}
```

### `/health`
`GET` – Returns a placeholder health object. Currently reports
`{"status":"unimplemented","db":"unimplemented"}`【F:internal/api/health.go†L9-L28】.

### `/ledger/{ledger}`
`GET` – Returns a ledger snapshot between optional `since` and `until`
(`YYYY-MM-DD`).【F:internal/api/ledger.go†L21-L70】

### `/ledger/{ledger}/transactions`
`GET` – List transactions with optional `limit` and `offset` query parameters.
`POST` – Create a transaction. Requires Bearer token and JSON body with
`date` (RFC3339), `amount` and `label`【F:internal/api/ledger.go†L71-L157】.

### `/metrics`
`GET` – Returns summary metrics about active patrons and revenue
【F:internal/api/metrics.go†L10-L24】.

### `/patrons`
`GET` – List patrons. Requires Bearer token. Supports `limit` and `offset`
parameters【F:internal/api/patrons.go†L11-L51】.

### `/settings/allocations`
`GET` – Retrieve allocation rules.
`PUT` – Replace all allocation rules. Body is an array of
`{id, ledger, percentage}` objects and percentages must total 100. Requires
Bearer token【F:internal/api/allocations.go†L12-L51】.

### `/settings/keys`
`POST` – Create a new API key. Returns the token once.
`DELETE /settings/keys/{id}` – Delete an API key. Both require a Bearer token
【F:internal/api/keys.go†L11-L45】.

### `/stripe/webhook`
`POST` – Endpoint for Stripe webhooks. Signature verified using the configured
secret. Returns `200 OK` on success, otherwise `400 Bad Request`
【F:internal/api/stripe.go†L12-L41】.

Authentication is handled via the `Authorization: Bearer <token>` header on
state‑changing routes. Tokens are issued by `/settings/keys` and validated on
each request.

## Running Tests

Execute all unit tests with:

```sh
$ go test ./...
```

All current tests pass:

```
ok      git.sr.ht/~jakintosh/coffer/internal/api        0.180s
ok      git.sr.ht/~jakintosh/coffer/internal/database   0.066s
ok      git.sr.ht/~jakintosh/coffer/internal/service    0.099s
```
【d09413†L1-L5】

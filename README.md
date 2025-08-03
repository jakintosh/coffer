# Coffer

## General Overview

Coffer is a small Go service that stores information about paying patrons and ledger balances in a local SQLite database. The command line tool (`coffer`) communicates with the HTTP server (started with `coffer serve`) to post ledger transactions, manage API keys and allocation rules, and to fetch summary metrics. The server also processes Stripe webhooks to keep patron records and payments in sync. All operations that modify state require an API key.

## Important Design Elements

- **Environment driven configuration** - Paths to the database file and listen port are taken from environment variables (`DB_FILE_PATH` and `PORT`). Additional credentials such as the Stripe key, webhook secret and bootstrap API key are loaded from files under a directory specified by `CREDENTIALS_DIRECTORY`. This conforms with the way `systemd` exposes credentials to services. `systemd` unit files are provided out of the box in the `init` folder.
- **Pluggable storage via interfaces** - The `service` package exposes interfaces for the ledger, patrons, allocation rules, metrics and Stripe events. Actual persistence uses the `internal/database` package, but the design allows other storage layers.
- **SQLite schema initialization** - On startup the server opens the database and creates tables for customers, subscriptions, payments, payouts, transactions, allocation rules and API keys if they do not already exist. Default allocation rules are inserted when none are present.
- **API key management** - API tokens are salted and hashed in the database. A bootstrap key can be provided for first run. New keys are created and revoked through the `/settings/keys` endpoints.
- **CORS whitelist managment** - Cross-Origin Resource Sharing origins are stored in the database and managed via the `/settings/cors` API. The `CORS_ALLOWED_ORIGINS` environment variable seeds the table when empty.
- **Stripe integration** - Webhook payloads are validated using the Stripe signature secret. Events update the customer, subscription, payment and payout tables and post ledger entries for successful payments.
- **Allocation based ledger posting** - Each payment is split across one or more ledgers using configurable percentage rules. The rules must sum to 100 percent.
- **Authentication middleware** - Mutating endpoints require the `Authorization: Bearer` header. Tokens are verified against the stored API keys before the request is forwarded.


## HTTP API Reference

This document describes the HTTP endpoints implemented under the `/api/v1` prefix.
Every response follows the structure defined in `internal/api/api.go`:

```json
{
  "error": { "code": int, "message": string } | null,
  "data": <payload or null>
}
```

Status codes and payloads for each route are listed below.

### `/health`
#### GET
Checks basic application status.

**Response Codes**
- `200 OK` – service and database reachable
- `503 Service Unavailable` – database check failed

**Response Body** ([`HealthResponse`](internal/api/health.go))
```json
{
  "status": "ok",
  "db": "ok" | "unreachable"
}
```

### `/ledger/{ledger}`
#### GET
Retrieve a ledger snapshot.

Path parameter `ledger` is the ledger name.

**Query Parameters**
- `since` (YYYY-MM-DD, optional) – start date. Defaults to epoch start.
- `until` (YYYY-MM-DD, optional) – end date. Defaults to current time.

Invalid dates return `400 Bad Request`.

**Response Codes**
- `200 OK` with snapshot
- `500 Internal Server Error` on storage errors

**Response Body** ([`LedgerSnapshot`](internal/service/ledger.go))
```json
{
  "opening_balance": int,
  "incoming_funds": int,
  "outgoing_funds": int,
  "closing_balance": int
}
```

#### `/ledger/{ledger}/transactions`
##### GET
List transactions for the ledger.

**Query Parameters**
- `limit` (integer, optional, default 100)
- `offset` (integer, optional, default 0)

Non‑integer values return `400 Bad Request`.

**Response Codes**
- `200 OK` with list
- `500 Internal Server Error` on storage errors

**Response Body** – array of [`Transaction`](internal/service/ledger.go)
```json
[
  {
    "id": string,
    "ledger": string,
    "amount": int,
    "date": "RFC3339 timestamp",
    "label": string
  }
]
```

##### POST *(requires `Authorization` header)*
Create a new transaction.

**Request Body** ([`CreateTransactionRequest`](internal/api/ledger.go))
```json
{
  "id": string?,
  "date": "RFC3339",
  "amount": int,
  "label": string
}
```

**Response Codes**
- `201 Created` on success
- `400 Bad Request` for malformed JSON or invalid date
- `401 Unauthorized` for missing/invalid token
- `500 Internal Server Error` on storage errors

### `/metrics`
#### GET
Returns summary subscription metrics.

**Response Codes**
- `200 OK` with metrics
- `500 Internal Server Error` if metrics collection fails

**Response Body** ([`Metrics`](internal/service/metrics.go))
```json
{
  "patrons_active": int,
  "mrr_cents": int,
  "avg_pledge_cents": int,
  "payment_success_rate_pct": number,
}
```

### `/patrons`
#### GET *(requires `Authorization` header)*
List known patrons.

**Query Parameters**
- `limit` (integer, optional, default 100)
- `offset` (integer, optional, default 0)

Invalid values return `400 Bad Request`.

**Response Codes**
- `200 OK` with array
- `500 Internal Server Error` on storage errors

**Response Body** – array of [`Patron`](internal/service/patrons.go)
```json
[
  {
    "id": string,
    "name": string,
    "created_at": "RFC3339 timestamp",
    "updated_at": "RFC3339 timestamp"
  }
]
```

### `/settings/allocations`
#### GET
Retrieve ledger allocation rules.

**Response Codes**
- `200 OK` with rules
- `500 Internal Server Error` on retrieval error

**Response Body** – array of [`AllocationRule`](internal/service/allocations.go)
```json
[
  {
    "id": string,
    "ledger": string,
    "percentage": int
  }
]
```

#### PUT *(requires `Authorization` header)*
Replace all allocation rules.

**Request Body** – array of the same `AllocationRule` objects. Percentages must sum to 100.
```json
[
  {
    "id": string,
    "ledger": string,
    "percentage": int
  }
]
```

**Response Codes**
- `204 No Content` on success
- `400 Bad Request` for malformed JSON or invalid percentages
- `401 Unauthorized` for missing/invalid token
- `500 Internal Server Error` on storage error

### `/settings/cors`
#### GET *(requires `Authorization` header)*
Retrieve the list of allowed CORS origins.

**Response Codes**
- `200 OK` with origins
- `401 Unauthorized` if token missing/invalid
- `500 Internal Server Error` on retrieval error

**Response Body** – array of [`AllowedOrigin`](internal/service/cors.go)
```json
[
  {
    "url": string
  }
]
```

#### PUT *(requires `Authorization` header)*
Replace all allowed origins.

**Request Body** – array of `AllowedOrigin` objects. URLs must start with `http://` or `https://`.
```json
[
  {
    "url": string
  }
]
```

**Response Codes**
- `204 No Content` on success
- `400 Bad Request` for malformed JSON or invalid origins
- `401 Unauthorized` for missing/invalid token
- `500 Internal Server Error` on storage error

### `/settings/keys`
#### POST *(requires `Authorization` header)*
Create a new API key.

**Response Codes**
- `201 Created` with generated token
- `401 Unauthorized` if token missing/invalid
- `500 Internal Server Error` on failure

**Response Body**
String token value (returned once).

### `/settings/keys/{id}` *(requires `Authorization` header)*
#### DELETE
Delete an API key by id.

**Response Codes**
- `204 No Content` on success
- `400 Bad Request` if id is empty
- `401 Unauthorized` if token invalid
- `500 Internal Server Error` on failure

### `/stripe/webhook`
#### POST
Stripe webhook endpoint. Payload is validated using the `Stripe-Signature` header. Only intended to be called by Stripe's API.

**Headers**
- `Stripe-Signature`: signature provided by Stripe

**Response Codes**
- `200 OK` when event accepted
- `400 Bad Request` if signature verification or body parsing fails

Body content is ignored; no data returned.

### Authentication
Endpoints that modify server state require an API key. Provide it via the `Authorization` header. Either `Bearer <token>` or just the raw token are accepted by the middleware implemented in [`middleware.go`](internal/api/middleware.go).


## CLI Usage

The `coffer` CLI separates remote API calls under the `api` command and
local configuration under `env`. Settings are stored in a `config.json`
file under the configuration directory (default `~/.config/coffer`).

```sh
# list configured environments ("*" marks the active one)
coffer env list

# create a new environment and bootstrap an API key
coffer env create staging --base-url http://staging.example.com --bootstrap

# switch the CLI to use that environment
coffer env activate staging

# delete an environment
coffer env delete staging

# post a transaction to a ledger
coffer api ledger tx create main --amount 1000 --date 2024-05-01T00:00:00Z --label example

# show current status
coffer status
```


## Running Tests

Run all unit tests with:

```sh
go test ./...
```

All current tests should pass.

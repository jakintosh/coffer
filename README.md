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


## HTTP API

See [API](./API.md)


## Running Tests

Run all unit tests with:

```sh
go test ./...
```

All current tests should pass.


## CLI Environments

The `coffer` CLI stores all of its settings in a single `config.json`
file under the configuration directory (default `~/.config/coffer`).
This file tracks the active environment and, for each environment, the
API key and base URL. Use `coffer auth env` to manage the entries:

```sh
# list configured environments ("*" marks the active one)
coffer auth env list

# create a new environment and bootstrap an API key
coffer auth env create staging --base-url http://staging.example.com \
  --bootstrap

# switch the CLI to use that environment
coffer auth env use staging

# delete an environment
coffer auth env delete staging
```

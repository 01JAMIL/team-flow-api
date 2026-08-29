# Gin API — Team Flow Backend

A production-style **REST API + WebSocket backend** for a project-management / team-collaboration SaaS ("Team Flow"), built with **Go**, the **Gin** framework, **PostgreSQL** (via **pgx + sqlc**), and integrated with **Stripe** for billing and **Resend** for transactional email.

It covers the full backend lifecycle: authentication, workspaces, workspace members (with roles), projects, tasks, real-time messaging, GitHub repository integrations with webhook-driven issue sync, and a complete Stripe subscription/billing flow with webhooks and email notifications.

---

## Features

- **Authentication**
  - Register / login with `bcrypt`-hashed passwords.
  - JWT (`HS256`) tokens with 7-day expiry issued on register/login.
  - A global `AuthenticationMiddleware` protects every `/api/v1` route except auth, health, and webhooks.

- **Workspaces**
  - Full CRUD scoped to the logged-in user.
  - **Free-plan limit**: 1 workspace for free; creating more requires an active **PRO** subscription.
  - Workspace creation runs inside a **transaction** that also adds the creator to `workspace_members` as **ADMIN**.

- **Workspace Members (RBAC)**
  - Add / list / remove members with a role of `ADMIN` or `MEMBER`.
  - Members list is paginated and joins user profile data.
  - Creating, updating, or deleting a **project** requires the caller to be an **ADMIN** member of the workspace (`FORBIDDEN` otherwise).

- **Projects**
  - CRUD nested under workspaces, paginated.
  - **Free-plan limit**: 2 projects per workspace; more requires an active PRO subscription.

- **Tasks**
  - CRUD nested under projects, paginated.
  - Fields: name, description, start/end dates (`YYYY-MM-DD`), status (`TODO | IN_PROGRESS | TESTING | DONE`), priority (`LOW | MEDIUM | HIGH | URGENT`), and an optional assignee.
  - Partial updates via `COALESCE`-based SQL (only provided fields are changed).

- **Real-time Messaging**
  - Messages between users over **WebSocket** (`/api/v1/ws`).
  - Sending a message to yourself is rejected (`SELF_MESSAGE_NOT_ALLOWED`).
  - Message history endpoint `GET /api/v1/messages/:userId` with **pagination** (newest first, both directions).

- **GitHub Integrations**
  - Connect a project to a GitHub repository (`username/repo_name`), verifying the repo via the GitHub API.
  - Each integration stores its own **webhook secret** (`crypto/rand`), used to verify incoming GitHub webhook payloads (HMAC-SHA256).
  - `GET /projects/:projectID/integrations/github` returns the connected integration; `POST .../regenerate-secret` rotates the webhook secret (**ADMIN only**).
  - GitHub **issue** webhooks are parsed into `integration_tasks`: an issue `opened` creates a task (status `open`), and `closed` updates it to status `closed`.

- **Billing & Subscriptions (Stripe)**
  - Checkout session creation per workspace (`POST /workspaces/:id/checkout`).
  - Lazy **Stripe Customer** creation, stored on the workspace.
  - Stripe **webhooks**: `checkout.session.completed`, `customer.subscription.updated`, `customer.subscription.deleted`, `invoice.payment_failed`.
  - Subscription plans: `PRO`, `BUSINESS`, `ENTERPRISE`; statuses `ACTIVE` / `INACTIVE`.
  - `GET /workspaces/:id/subscription` returns the workspace's active subscription.
  - Payment-failure notifications via Resend email.

- **Consistent API design**
  - Every list endpoint returns a `{ ..., "pagination": {...} }` envelope (page, pageSize, total, totalPages).
  - Every error has a uniform shape `{ "code", "message", "errors?" }` handled centrally by `codeerror`.

---

## Tech Stack

| Layer      | Technology |
|------------|------------|
| Language   | Go 1.26 |
| HTTP       | [Gin](https://github.com/gin-gonic/gin) v1.12 |
| Database   | PostgreSQL 16 (Docker Compose) |
| DB driver  | [pgx/v5](https://github.com/jackc/pgx) |
| SQL layer  | [sqlc](https://sqlc.dev) (type-safe generated Go) |
| Migrations | [goose](https://github.com/pressly/goose) (SQL files) |
| Auth       | `golang-jwt/jwt/v5` + `golang.org/x/crypto/bcrypt` |
| Realtime   | `gorilla/websocket` |
| Payments   | `stripe/stripe-go/v86` |
| Email      | `resend/resend-go/v3` |
| Config     | `godotenv` + `internal/env` helpers |
| UUIDs      | `google/uuid` |

---

## Project Structure

```
gin-api-1/
├── cmd/api/                      # Application entrypoint
│   ├── main.go                   # Env, DB connection, clients (Stripe/Resend), bootstrap
│   ├── serve.go                  # HTTP server config (timeouts, port)
│   └── routes.go                 # All route registration + dependency wiring
├── internal/
│   ├── adapters/postgresql/
│   │   ├── migrations/           # goose SQL migrations 00001–00012
│   │   └── sqlc/                 # generated repo (models.go, queries.sql.go, db.go, querier.go)
│   ├── auth/                     # register/login, JWT, bcrypt, auth middleware
│   ├── codeerror/                # centralized error codes + HTTP status mapping
│   ├── email/                    # Resend email service (welcome, payment failed)
│   ├── env/                      # environment variable helpers
│   ├── integrations/             # GitHub repo connect/status/regenerate + webhook issue sync
│   ├── messages/                 # messaging service + handlers + types
│   ├── payment/                  # Stripe service + webhook handler
│   ├── projects/                 # project CRUD (ADMIN-gated, free-plan limits)
│   ├── subscriptions/            # subscription service + handler + types
│   ├── tasks/                    # task CRUD
│   ├── websocket/                # WS hub, client read/write pumps, message broadcasting
│   ├── workspace/                # workspace CRUD + checkout session
│   └── workspacemembers/         # member add/list/remove
├── client/client.html            # tiny WebSocket test page
├── docker-compose.yaml           # local PostgreSQL 16
├── sqlc.yaml                     # sqlc config
├── Makefile                      # run / build / docker-up / docker-down
└── go.mod
```

Each domain package follows a consistent **handler → service → types** layout:

- `types.go` — request payloads, response DTOs, and `DefaultPageSize = 10`.
- `services.go` — business logic, validation, DB access via the generated repo.
- `handlers.go` — HTTP handlers; parse params/body, call the service, render JSON.

---

## Getting Started

### Prerequisites

- Go 1.26+
- Docker (for PostgreSQL)
- [sqlc](https://sqlc.dev) — only needed if you modify SQL queries
- [goose](https://github.com/pressly/goose) — to run migrations
- Stripe + Resend accounts (optional for a full local test of billing/email)

### 1. Start PostgreSQL

```bash
make docker-up
```

This starts `postgres:16-alpine` on `localhost:5432` (database `gin-api-1`, user/password `postgres/postgres` unless overridden).

### 2. Configure environment

Copy the following into a `.env` file at the repo root:

```env
PORT=3700

# PostgreSQL DSN used by the API
GOOSE_DBSTRING=host=localhost user=postgres password=postgres dbname=gin-api-1 sslmode=disable

# Secrets (change these!)
JWT_SECRET=change_me

# Base URL used to build the GitHub webhook URL returned when connecting a repository
APP_BASE_URL=http://localhost:3700

# GitHub (optional but recommended) — used to verify repositories via the GitHub API without hitting rate limits
GITHUB_TOKEN=

# Stripe
STRIPE_SECRET_KEY=sk_test_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx
STRIPE_PRO_PRICE_ID=price_xxx
STRIPE_SUCCESS_URL=http://localhost:3000/payment/success
STRIPE_CANCEL_URL=http://localhost:3000/payment/cancel

# Resend (transactional email)
RESEND_API_KEY=re_xxx
```

> `docker-compose.yaml` additionally reads `DATABASE_USERNAME`, `DATABASE_PASSWORD`, `DATABASE_NAME` with defaults `postgres` / `mypassword` / `gin-api-1`. Set these to match your `GOOSE_DBSTRING` if you customize them.

### 3. Run migrations

```bash
goose -dir internal/adapters/postgresql/migrations postgres "$GOOSE_DBSTRING" up
```

### 4. Run the API

```bash
make run            # go run ./cmd/api
# or
make build          # go build -o bin/api ./cmd/api
```

The server logs `Starting server on port 3700` (or whatever `PORT` is set to).

### Regenerating SQL code

When you edit `queries.sql` or the migrations:

```bash
sqlc generate
```

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | no | `8080` | HTTP port the API listens on |
| `GOOSE_DBSTRING` | yes* | `host=localhost user=postgres password=postgres dbname=postgres sslmode=disable` | PostgreSQL connection DSN |
| `JWT_SECRET` | no* | `secret_123456` | Secret used to sign JWT tokens |
| `APP_BASE_URL` | no | `http://localhost:3700` | Base URL used to build the GitHub webhook URL in integration responses |
| `GITHUB_TOKEN` | no | `""` | GitHub personal access token for the GitHub API (avoids unauthenticated rate limits when verifying repositories) |
| `STRIPE_SECRET_KEY` | yes* | `stripe_xx` | Stripe API secret key |
| `STRIPE_WEBHOOK_SECRET` | no | `""` | Signature verification for webhooks |
| `STRIPE_PRO_PRICE_ID` | no | `price_xx` | Price ID used for the PRO checkout |
| `STRIPE_SUCCESS_URL` | no | `http://localhost:3000/payment/success` | Checkout success redirect |
| `STRIPE_CANCEL_URL` | no | `http://localhost:3000/payment/cancel` | Checkout cancel redirect |
| `RESEND_API_KEY` | no* | `re_xx` | Resend API key for email |

\* Required for real functionality; the API has safe defaults so it boots without them.

---

## API Reference

All endpoints are prefixed with `/api/v1`. Authenticated endpoints require:

```
Authorization: Bearer <token>
```

### Public

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check → `{ "message": "OK" }` |
| `POST` | `/auth/register` | Create an account + JWT |
| `POST` | `/auth/login` | Login + JWT |
| `POST` | `/webhooks/stripe` | Stripe webhook receiver (no auth) |
| `POST` | `/webhooks/github` | GitHub webhook receiver — syncs issue events into `integration_tasks` (no auth) |

#### Register — `POST /auth/register`

```json
{
  "firstName": "Jamil",
  "lastName": "Ben",
  "email": "jamil@example.com",
  "password": "password123"
}
```

Response `201`:

```json
{
  "message": "User created successfully",
  "user": {
    "user": { "id": "...", "firstName": "Jamil", "lastName": "Ben", "email": "...", "createdAt": "...", "updatedAt": null },
    "token": "<jwt>"
  }
}
```

> Registration also fires a Resend welcome email (errors are logged, not fatal).

#### Login — `POST /auth/login`

```json
{ "email": "jamil@example.com", "password": "password123" }
```

Response `200` has the same `{ "message", "user": { "user", "token" } }` shape.

#### Get current user — `GET /auth/me`

```json
{ "message": "User retrieved successfully", "user": { "id": "...", "firstName": "Jamil", ... } }
```

### Workspaces

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/workspaces` | List current user's workspaces (paginated) |
| `GET` | `/workspaces/:id` | Get one of the user's workspaces |
| `POST` | `/workspaces` | Create a workspace (creator becomes ADMIN member) |
| `PATCH` | `/workspaces/:id` | Update workspace name/description |
| `DELETE` | `/workspaces/:id` | Delete a workspace |
| `POST` | `/workspaces/:id/checkout` | Create a Stripe checkout session for this workspace |

#### Create — `POST /workspaces`

```json
{ "workspaceName": "SupaGo", "description": "Go backend team" }
```

Response `201`:

```json
{
  "message": "Workspace created successfully",
  "workspace": {
    "id": "d949ce46-...", "workspace_name": "SupaGo", "description": "Go backend team",
    "user_id": "...", "created_at": "...", "updated_at": null, "stripe_customer_id": null
  }
}
```

> Single-workspace GET/PATCH/DELETE return the raw `workspace` row; the list endpoint returns the `{ workspaces, pagination }` envelope.
>
> **Limit logic:** workspaces are only created if the user has **zero workspaces** (free tier) **or** an active **PRO** subscription. Creating your 1st workspace happens in a transaction that inserts the workspace **and** a `workspace_members` row with `user_role = ADMIN` for the caller.

### Workspace Members

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/workspaces/:id/members` | List members (paginated, joined with user info) |
| `POST` | `/workspaces/:id/members` | Add a member (`ADMIN` or `MEMBER`) |
| `DELETE` | `/workspaces/:id/members/:userId` | Remove a member |

#### Add member — `POST /workspaces/:id/members`

```json
{ "userId": "uuid-of-user", "userRole": "MEMBER" }
```

Response `201`:

```json
{
  "message": "Member added successfully",
  "member": { "id": "...", "user_id": "...", "workspace_id": "...", "user_role": "MEMBER", "created_at": "..." }
}
```

#### List members — `GET /workspaces/:id/members`

```json
{
  "workspace": { "id": "...", "workspaceName": "SupaGo", "description": "Go backend team", "userId": "...", "createdAt": "...", "updatedAt": null },
  "members": [
    { "id": "...", "userId": "...", "workspaceId": "...", "userRole": "MEMBER", "createdAt": "...",
      "user": { "id": "...", "firstName": "Jamil", "lastName": "Ben", "email": "jamil@example.com", "createdAt": "...", "updatedAt": null } }
  ],
  "pagination": { "page": 1, "pageSize": 10, "total": 1, "totalPages": 1 }
}
```

### Projects

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/workspaces/:id/projects` | List workspace projects (paginated) |
| `POST` | `/workspaces/:id/projects` | Create a project (**ADMIN only**) |
| `GET` | `/projects/:projectID` | Get one project |
| `PATCH` | `/projects/:projectID` | Update a project (**ADMIN only**) |
| `DELETE` | `/projects/:projectID` | Delete a project (**ADMIN only**) |

#### Create — `POST /workspaces/:id/projects`

```json
{ "name": "Team Flow", "description": "Core product" }
```

Response `201`:

```json
{ "message": "Project created successfully", "project": { "id": "...", "name": "Team Flow", "description": "Core product", "workspaceId": "...", "createdAt": "...", "updatedAt": null } }
```

> **Authorization:** project create/update/delete require the caller to be an `ADMIN` workspace member (`MEMBER_NOT_FOUND` / `FORBIDDEN` otherwise).
> **Limit logic:** free tier allows **2 projects** per workspace; beyond that an active **PRO** subscription is required (`FREE_PLAN_PROJECT_LIMIT_REACHED`).

### Tasks

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/projects/:projectID/tasks` | Create a task |
| `GET` | `/projects/:projectID/tasks` | List project tasks (paginated) |
| `GET` | `/tasks/:id` | Get one task |
| `PATCH` | `/tasks/:id` | Update task (partial) |
| `DELETE` | `/tasks/:id` | Delete a task |

#### Create — `POST /projects/:projectID/tasks`

```json
{
  "name": "Design database schema",
  "description": "Define tables and relations",
  "startDate": "2026-08-01",
  "endDate": "2026-08-05",
  "status": "TODO",
  "priority": "HIGH",
  "assigneeId": "uuid-of-user"
}
```

Response `201`:

```json
{
  "message": "Task created successfully",
  "task": {
    "id": "...", "name": "Design database schema", "description": "Define tables and relations",
    "startDate": "2026-08-01", "endDate": "2026-08-05", "status": "TODO", "priority": "HIGH",
    "projectId": "...", "assigneeId": "uuid-of-user", "createdAt": "...", "updatedAt": null
  }
}
```

> Dates are parsed as `YYYY-MM-DD` (`INVALID_DATE` on bad input). Updates only touch fields you include (nil-coalesced SQL).

### Integrations (GitHub)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/projects/:projectID/integrations/github` | Get the project's connected repository |
| `POST` | `/projects/:projectID/integrations/github` | Connect a repository to the project (**ADMIN only**) |
| `POST` | `/projects/:projectID/integrations/github/regenerate-secret` | Rotate the webhook secret (**ADMIN only**) |

#### Get integration — `GET /projects/:projectID/integrations/github`

```json
{
  "integration": {
    "id": "...", "projectId": "...", "provider": "github",
    "repositoryOwner": "01JAMIL", "repositoryName": "team-flow",
    "isActive": true, "createdAt": "...", "updatedAt": "..."
  }
}
```

> Returns `PROJECT_NOT_FOUND` if no repository is connected to the project. The `webhookSecret` is deliberately **not** exposed here.

#### Connect repository — `POST /projects/:projectID/integrations/github`

```json
{ "repository": "01JAMIL/team-flow" }
```

Response `201`:

```json
{
  "message": "Repository connected successfully",
  "integration": {
    "provider": "github",
    "repository": "01JAMIL/team-flow",
    "webhookUrl": "https://your-domain/api/v1/webhooks/github",
    "webhookSecret": "...generated..."
  }
}
```

> **Flow:** validates the `owner/repo_name` format → loads the project → checks the caller is an **ADMIN** workspace member (`FORBIDDEN` otherwise) → ensures the project isn't already connected (`INTEGRATION_ALREADY_EXISTS`) → verifies the repository exists via `https://api.github.com/repos/{owner}/{repo_name}` (`REPOSITORY_NOT_FOUND`) → generates a `crypto/rand` webhook secret → persists the integration.
>
> Point the GitHub webhook at the returned `webhookUrl` (Content type `application/json`, select the **Issues** event) and set the **Secret** to `webhookSecret`.

#### Regenerate secret — `POST /projects/:projectID/integrations/github/regenerate-secret`

Response `200`:

```json
{
  "message": "Webhook secret regenerated successfully",
  "integration": {
    "provider": "github",
    "repository": "01JAMIL/team-flow",
    "webhookUrl": "https://your-domain/api/v1/webhooks/github",
    "webhookSecret": "...newly generated..."
  }
}
```

> **ADMIN only.** Rotates the stored webhook secret, so any GitHub webhook must be reconfigured with the new value.

### Messaging

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/messages/:userId` | Conversation between the logged-in user and `:userId` (paginated, newest first) |

#### List conversation — `GET /messages/:userId?page=1&pageSize=10`

```json
{
  "messages": [
    { "id": "...", "senderId": "...", "receiverId": "...", "content": "Hey!", "createdAt": "2026-08-13T12:00:00Z" }
  ],
  "pagination": { "page": 1, "pageSize": 10, "total": 1, "totalPages": 1 }
}
```

> The conversation includes messages sent **in both directions** between the two users. Sending messages goes through the WebSocket channel (below).

### Subscriptions (Billing)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/workspaces/:id/subscription` | Get the workspace's active subscription |
| `POST` | `/workspaces/:id/checkout` | Start a Stripe checkout for a PRO subscription |

#### Get subscription — `GET /workspaces/:id/subscription`

```json
{
  "subscription": {
    "id": "...", "workspaceId": "...", "stripeSubscriptionId": "sub_xxx", "stripePriceId": "price_xxx",
    "status": "ACTIVE", "plan": "PRO",
    "currentPeriodStart": "2026-08-13T00:00:00Z", "currentPeriodEnd": "2026-09-13T00:00:00Z",
    "createdAt": "...", "updatedAt": "..."
  }
}
```

> Returns `WORKSPACE_NOT_FOUND` if the workspace doesn't belong to the user, or `SUBSCRIPTION_NOT_FOUND` if no active subscription exists.

---

## Pagination

Every **list** endpoint accepts the same query params and returns the same envelope:

```
?page=1&pageSize=10
```

- `page` defaults to `1`, `pageSize` defaults to `10`; invalid values fall back to the defaults.
- The response includes `pagination`: `{ "page", "pageSize", "total", "totalPages" }`.
- `total`/`totalPages` come from a `count(*) OVER ()` window column added to the SQL query, so only one DB round-trip is needed.

---

## Error Handling

All errors are rendered centrally by `codeerror.HandleError` with a uniform shape:

```json
{
  "code": "PROJECT_NOT_FOUND",
  "message": "Project not found",
  "errors": { "field": "validation detail" }   // only on validation errors
}
```

Common codes:

| Code | HTTP |
|------|------|
| `UNAUTHORIZED`, `INVALID_TOKEN`, `MISSING_TOKEN`, `INVALID_CREDENTIALS` | 401 |
| `FORBIDDEN`, `FREE_PLAN_WORKSPACE_LIMIT_REACHED`, `FREE_PLAN_PROJECT_LIMIT_REACHED`, `PRO_PLAN_REQUIRED` | 403 |
| `NOT_FOUND`, `WORKSPACE_NOT_FOUND`, `PROJECT_NOT_FOUND`, `TASK_NOT_FOUND`, `USER_NOT_FOUND`, `MEMBER_NOT_FOUND`, `SUBSCRIPTION_NOT_FOUND`, `REPOSITORY_NOT_FOUND` | 404 |
| `CONFLICT`, `USER_ALREADY_EXIST`, `MEMBER_ALREADY_EXISTS`, `INTEGRATION_ALREADY_EXISTS` | 409 |
| `BAD_REQUEST`, `VALIDATION_ERROR`, `INVALID_UUID`, `INVALID_DATE`, `SELF_MESSAGE_NOT_ALLOWED`, `FAILED_TO_DEACTIVATE_SUBSCRIPTION`, `INVALID_REPOSITORY_FORMAT` | 400 |
| anything else / raw errors | 500 (logged, generic message) |

---

## WebSocket Realtime Messaging

Connect to `GET /api/v1/ws` (authenticated) and send a JSON message:

```json
{ "receiverId": "uuid-of-receiver", "content": "Hello there!" }
```

- The server validates the payload, persists the message, and **broadcasts** the created `MessageResponse` to both the sender and the receiver (if connected).
- Sending to yourself returns an error frame: `{"error":"..."}`.
- The hub keeps a single connection per user ID (`clients map[string]*Client`); one client per user.
- A quick test page lives at `client/client.html`.

---

## Stripe Billing Flow

1. **Checkout** — `POST /workspaces/:id/checkout`:
   - If the workspace has no Stripe customer yet, a customer is created and saved via `UpdateWorkspaceStripeCustomer`.
   - Returns `{ "url": "<checkout_session_url>" }` for the PRO price (`STRIPE_PRO_PRICE_ID`).
   - The checkout session carries `workspace_id` in its metadata.

2. **Webhook: `checkout.session.completed`** — reads `workspace_id` metadata, retrieves the Stripe subscription, and inserts a `subscriptions` row (`status=ACTIVE`, `plan=PRO`, current period = now → +1 month).

3. **Webhook: `customer.subscription.updated`** — updates the subscription's price/status/period.

4. **Webhook: `customer.subscription.deleted`** — sets the subscription `status=INACTIVE`.

5. **Webhook: `invoice.payment_failed`** — looks up the user by the customer's Stripe ID (`GetUserByStripeCustomerID`) and sends a Resend "Payment Failed" email.

Webhook payloads are verified against `STRIPE_WEBHOOK_SECRET` with `webhook.ConstructEventWithOptions`.

---

## GitHub Webhook Flow

1. **Connect** — `POST /projects/:projectID/integrations/github` verifies the repo on GitHub and stores the integration with a per-project `webhook_secret`. The response exposes the webhook URL (`{APP_BASE_URL}/api/v1/webhooks/github`) you configure in the repository's **Settings → Webhooks**.

2. **Webhook: issue `opened`** — GitHub signs the payload with HMAC-SHA256 using the integration's `webhook_secret`. The handler looks up the integration by `repository/repository_name` (`PROJECT_NOT_FOUND` / 404 if nothing is connected), recomputes and compares the `X-Hub-Signature-256` header (`UNAUTHORIZED` / 401 on mismatch), then inserts an `integration_tasks` row with `status='open'`.

3. **Webhook: issue `closed`** — same verification; the matching `integration_tasks` row (by `external_id`) is updated to `status='closed'`.

4. **Regenerate** — `POST .../regenerate-secret` rotates the `webhook_secret`; point the GitHub webhook's **Secret** field at the new value.

Outgoing calls to `api.github.com` include `User-Agent: gin-api` and (when `GITHUB_TOKEN` is set) an `Authorization: Bearer` header so unauthenticated rate limits don't break repository verification.

---

## Database Schema

Ten tables (migrations `00001`–`00012`), all keyed by **UUID** primary keys:

```
users (id, first_name, last_name, email UNIQUE, password, created_at, updated_at)
workspaces (id, workspace_name, description, user_id FK→users, created_at, updated_at, stripe_customer_id)
workspace_members (id, user_id FK→users, workspace_id FK→workspaces, user_role CHECK IN ('ADMIN','MEMBER'), created_at)
projects (id, name, description, workspace_id FK→workspaces, created_at, updated_at)
tasks (id, name, description, start_date, end_date,
       status CHECK IN ('TODO','IN_PROGRESS','TESTING','DONE'),
       priority CHECK IN ('LOW','MEDIUM','HIGH','URGENT'),
       project_id FK→projects, assignee_id FK→users, created_at, updated_at)
messages (id, sender_id FK→users, receiver_id FK→users, content, created_at)
subscriptions (id, workspace_id FK→workspaces, stripe_subscription_id UNIQUE, stripe_price_id,
               status CHECK IN ('ACTIVE','INACTIVE'), plan CHECK IN ('PRO','BUSINESS','ENTERPRISE'),
               current_period_start, current_period_end, created_at, updated_at)
project_integrations (id, project_id FK→projects ON DELETE CASCADE, provider CHECK IN ('github','gitlab','jira'),
                      repository_owner, repository_name, webhook_secret, is_active DEFAULT TRUE,
                      created_at, updated_at, UNIQUE (provider, repository_owner, repository_name))
integration_tasks (id, project_id FK→projects ON DELETE CASCADE, provider CHECK IN ('github','gitlab','jira'),
                   resource_type CHECK IN ('issue','pull_request','ticket','merge_request'),
                   external_id UNIQUE, repository_name, issue_number, title, description,
                   status CHECK IN ('open','closed'), assignee_id FK→users ON DELETE SET NULL,
                   payload JSONB, created_at, updated_at)
```

Migrations `00009`/`00010`/`00012` add the integration tables; `00011` indexes `project_integrations` on `(provider, repository_owner, repository_name)` for fast webhook lookup.

---

## Useful Commands

```bash
make run          # run the API
make build        # build ./bin/api
make test         # run all unit tests (verbose)
make docker-up    # start PostgreSQL
make docker-down  # stop PostgreSQL

sqlc generate     # regenerate SQL layer after editing queries
goose -dir internal/adapters/postgresql/migrations postgres "$GOOSE_DBSTRING" up    # apply migrations
goose -dir internal/adapters/postgresql/migrations postgres "$GOOSE_DBSTRING" down  # rollback migrations
```

---

## Notes

- The app is intended for **learning/portfolio** purposes — it demonstrates clean layering (handlers → services → generated repo), centralized error handling, RBAC, free/pro plan gating, transactions, pagination, an interface-based service layer for testability, and a WebSocket hub. Unit tests live alongside the services under `internal/auth`, `internal/messages`, and `internal/integrations` (run `make test` or `go test ./...`). For production, you'd add request timeouts/rate limiting, structured logging, connection pooling (`pgxpool`), refresh tokens, and integration tests against a real database.

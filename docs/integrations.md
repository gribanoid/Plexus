# Platform extensibility & integrations

Plexus extends via **stable HTTP API + outbound webhooks + automation rules** inside the modular monolith — not by splitting the core into microservices.

## Public API

- Contract: [`backend/openapi.yaml`](../backend/openapi.yaml)
- Auth: Bearer JWT or `X-API-Key` (org API keys with scopes)
- Version prefix: `/api/v1`

Build bots and marketplace apps against OpenAPI; prefer codegen for typed clients.

## Outbound webhooks

Org admins register webhook endpoints. On events (issue created/updated, transition, comment, sprint), Plexus POSTs a signed JSON payload:

```json
{
  "id": "evt_…",
  "type": "issue.updated",
  "org_id": "…",
  "project_id": "…",
  "created_at": "…",
  "data": { }
}
```

Headers:

- `X-Plexus-Signature`: `sha256=<hmac of body with webhook secret>`
- `X-Plexus-Event`: event type
- `X-Plexus-Delivery`: delivery UUID

Deliveries are queued on asynq (`webhook:deliver`) and retried with backoff.

## Automation rules (v1)

Project rules: `when` trigger + `if` conditions + `then` actions (assign, transition, add label, webhook). Evaluated in-process after issue mutations; heavy side-effects go through the worker.

## Recommended first integrations

| Integration | Pattern |
|-------------|---------|
| GitHub / GitLab | Link MRs via issue key in branch/title; status webhook → comment |
| Slack / chat | Outbound webhook → channel notify on assign/mention |
| CI | API key + transition issue on deploy |
| Import | CSV bulk endpoints |

## Marketplace direction

Apps are external processes that:

1. Use OAuth/OIDC or API keys,
2. Subscribe to webhooks,
3. Call `/api/v1` for mutations.

No in-process plugin ABI in v1; keep the core binary simple for self-host.

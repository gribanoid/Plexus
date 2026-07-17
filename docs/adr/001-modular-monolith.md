# ADR 001: Modular monolith (not microservices)

## Status

Accepted

## Context

Plexus is a self-hosted enterprise project management product. A common early question is whether to split the Go backend into domain microservices (issues, projects, auth, search, etc.).

A Go modular monolith can scale to tens/hundreds of thousands of users via clustered replicas sharing Postgres, Redis/cache, and object storage — without domain service meshes.

## Decision

1. **Keep a single primary deployable** (`plexus-server`) that owns HTTP API, WebSocket hub wiring, and domain logic.
2. **Allow an optional second process** (`plexus-worker`) for asynq jobs (email, search indexing, webhooks) using the **same codebase and image**, different entrypoint — not a domain microservice.
3. **Enforce module boundaries in-process** (`internal/service`, repository interfaces) so code can later be extracted if load or team boundaries demand it.
4. **Do not** split domain services (issue-service, project-service, auth-service) or introduce separate databases per domain until there is proven scale pressure and independent team/SLA needs.

## Consequences

### Positive

- Simple self-host story (Compose / one binary) remains a product differentiator vs multi-service stacks.
- Faster feature delivery for workflows, RBAC, automation.
- Horizontal scale via stateless API replicas + shared infra.

### Negative / trade-offs

- Single binary can become large; discipline around packages is required.
- Independent deploy of “just search” is deferred until extraction is justified.

## When to revisit

Extract a process (still shared repo) when:

- Job/indexing load saturates API pods, or
- Mobile push requires an isolated proxy, or
- Search indexing must scale independently.

Extract a **domain** microservice only with measured bottlenecks, separate ownership, and clear network contracts.

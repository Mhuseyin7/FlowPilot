# API v1

The API uses JSON and returns an `X-Request-Id` response header.

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/healthz` | Liveness check |
| `GET` | `/api/v1/workflows` | List workflows |
| `POST` | `/api/v1/workflows` | Create a draft |
| `PUT` | `/api/v1/workflows/{id}/draft` | Validate and update a draft with its revision |
| `POST` | `/api/v1/workflows/{id}/publish` | Create and activate immutable version |
| `POST` | `/api/v1/workflows/{id}/executions` | Start a manual run |
| `GET` | `/api/v1/workflows/{id}/executions` | List execution history |
| `POST` | `/api/v1/executions/{id}/cancel` | Cancel a queued/running execution |
| `POST` | `/hooks/{webhookId}` | Enqueue an active webhook-trigger workflow |

Webhook trigger configuration requires a unique `webhookId`. Request bodies are JSON and limited to 1 MiB. Authentication and signature verification are not yet implemented; do not expose hooks publicly.

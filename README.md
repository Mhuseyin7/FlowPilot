# FlowPilot

> **Automate your workflows. Own your infrastructure.**

FlowPilot; API, webhook, scheduled task ve data transformation adımlarını görsel workflow'larda birleştirmek için geliştirilen self-hosted, open-source automation platformudur.

Bu project açık kaynak olarak herkese açıktır ve [muhammedkoca.com.tr](https://muhammedkoca.com.tr) tarafından geliştirilmiştir.

> **Project status:** Early development. Temel execution altyapısı vardır; production deployment öncesinde Roadmap'teki security ve product özellikleri tamamlanmalıdır.

## Why FlowPilot?

- **Self-hosted:** Infrastructure, database ve execution data sizin control'ünüzdedir.
- **Open-source:** Source code inceleyebilir, fork'layabilir ve contribution yapabilirsiniz.
- **Durable execution:** Execution state PostgreSQL'de tutulur; worker restart sonrası job kayıtları korunur.
- **Immutable versions:** Publish edilen workflow version sonradan değişmez.
- **Safe by default:** Arbitrary shell/JavaScript execution yoktur; HTTP node private/local hedefleri engeller.
- **API-first:** Workflow, draft, publish ve execution işlemleri REST API üzerinden yapılır.

## Features

| Area | Mevcut özellik |
| --- | --- |
| Workflow | Draft, optimistic revision, DAG validation, immutable published version |
| Execution | PostgreSQL durable jobs, worker lease, cancellation, execution events |
| Nodes | Manual/webhook/schedule trigger, JSON transform, set variable, condition, bounded delay, HTTP request, workflow result |
| API | Workflow create/list, publish, manual run, history, cancel ve webhook endpoint |
| Deployment | PostgreSQL, API, worker ve web için Docker Compose |
| UI | Başlangıç workflow list/create interface |

## Quick Start

### Requirements

- Docker Desktop / Docker Engine
- Docker Compose v2

### 1. Configure environment

```bash
cp .env.example .env
```

`.env` içindeki `POSTGRES_PASSWORD` ve `FLOWPILOT_ENCRYPTION_KEY` değerlerini güvenli değerlerle değiştirin. Example secret'ları production deployment'ta kullanmayın.

### 2. Start the stack

```bash
docker compose up --build
```

İlk empty PostgreSQL volume oluşturulduğunda migration otomatik uygulanır. Web interface `http://localhost:3000`, API ise `http://localhost:8080` adresinde açılır.

### 3. Create a workflow

```bash
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{"name":"Contact form workflow","description":"Example workflow"}'
```

Workflow graph örneği: [`examples/contact-form-to-discord.json`](examples/contact-form-to-discord.json).

## Webhook Example

Bir `webhook_trigger` node'una benzersiz `webhookId` verin:

```json
{"id":"trigger","type":"webhook_trigger","config":{"webhookId":"contact-form-demo"}}
```

Workflow publish edilip `ACTIVE` olduğunda:

```bash
curl -X POST http://localhost:8080/hooks/contact-form-demo \
  -H "Content-Type: application/json" \
  -d '{"email":"hello@example.com","message":"Merhaba FlowPilot"}'
```

API `202 Accepted` ve `executionId` döndürür. Worker publish edilmiş immutable graph'ı çalıştırır; node event'leri persist edilir.

> **Security notice:** Webhook authentication/HMAC verification henüz yoktur. Endpoint'i public internet'e açmadan önce reverse-proxy authentication kullanın veya signed-webhook feature'ını tamamlayın.

## Architecture

```text
Web UI → API Server → PostgreSQL ← Worker
                         ↑            │
                    Durable jobs ← execution events
```

PostgreSQL workflow definition, immutable version, execution ve job state için authoritative storage'dır. API long-running workflow'u request içinde çalıştırmaz; durable job oluşturur. Worker job'ı `FOR UPDATE SKIP LOCKED` ve lease modeliyle claim eder.

Detaylar: [Architecture](docs/ARCHITECTURE.md) · [API reference](docs/API.md) · [Configuration](docs/CONFIGURATION.md)

## API Overview

| Method | Endpoint | Açıklama |
| --- | --- | --- |
| `GET` | `/healthz` | Liveness check |
| `GET` | `/api/v1/workflows` | Workflow listesi |
| `POST` | `/api/v1/workflows` | Draft oluşturur |
| `PUT` | `/api/v1/workflows/{id}/draft` | Draft graph validate/save |
| `POST` | `/api/v1/workflows/{id}/publish` | Immutable version publish |
| `POST` | `/api/v1/workflows/{id}/executions` | Manual execution |
| `GET` | `/api/v1/workflows/{id}/executions` | Execution history |
| `POST` | `/api/v1/executions/{id}/cancel` | Execution cancel |
| `POST` | `/hooks/{webhookId}` | Webhook execution enqueue |

## Development

```bash
make dev
make test
make build
make migrate
```

Backend Go; frontend Next.js + TypeScript kullanır. CI, Go test/vet ve frontend typecheck/build adımlarını çalıştırır.

## Security

FlowPilot default olarak arbitrary code execution sunmaz. HTTP node loopback/private/link-local hedefleri DNS resolution aşamasında reddeder. Credential vault, RBAC, audit log, secret redaction, signed webhook ve rate limiting henüz tamamlanmamıştır.

Deployment öncesinde [Security status](docs/SECURITY.md) ve [Threat model](THREAT_MODEL.md) belgelerini mutlaka okuyun.

## Roadmap

- [ ] Owner setup, login, Argon2id session security ve organization RBAC
- [ ] Encrypted credential vault ve credential references
- [ ] Signed/HMAC webhook, replay protection ve rate limiting
- [ ] Timezone-aware scheduler ve interval trigger
- [ ] SMTP, Discord ve advanced HTTP configuration
- [ ] React Flow visual workflow editor ve execution inspector
- [ ] Import/export, templates, Turkish/English localization
- [ ] Integration, E2E, restart recovery ve security test suites
- [ ] OpenAPI, observability metrics ve release pipeline

## Contributing

Contribution'lar memnuniyetle karşılanır. Yeni feature için issue açın; küçük, test edilebilir pull request'ler tercih edilir. Commit'lere secret, API token veya real credential eklemeyin.

## License

FlowPilot [MIT License](LICENSE) ile lisanslanmıştır.

---

Made as an open-source project by [muhammedkoca.com.tr](https://muhammedkoca.com.tr).

# Configuration

| Variable | Required | Description |
| --- | --- | --- |
| `DATABASE_URL` | Yes | PostgreSQL connection URL used by API and worker. |
| `POSTGRES_DB` | Docker | Database name. |
| `POSTGRES_USER` | Docker | Database user. |
| `POSTGRES_PASSWORD` | Yes | Strong database password; never use the sample value. |
| `FLOWPILOT_API_ADDR` | No | API listen address; defaults to `:8080`. |
| `FLOWPILOT_ENCRYPTION_KEY` | Future | Reserved for the planned credential vault. |

The current Compose file initializes database schema only on an empty database volume. Back up PostgreSQL before replacing a volume.

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TYPE workflow_status AS ENUM ('DRAFT','ACTIVE','PAUSED','DISABLED','ARCHIVED');
CREATE TYPE execution_status AS ENUM ('QUEUED','STARTING','RUNNING','WAITING','RETRYING','SUCCEEDED','FAILED','CANCELLED','TIMED_OUT');
CREATE TABLE workflows (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), name TEXT NOT NULL CHECK (char_length(name) <= 160), description TEXT NOT NULL DEFAULT '', status workflow_status NOT NULL DEFAULT 'DRAFT', draft JSONB NOT NULL DEFAULT '{"nodes":[],"edges":[]}', draft_revision BIGINT NOT NULL DEFAULT 1, published_version INTEGER, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE workflow_versions (
 workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE, version INTEGER NOT NULL, definition JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), PRIMARY KEY(workflow_id, version)
);
CREATE TABLE executions (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), workflow_id UUID NOT NULL REFERENCES workflows(id), workflow_version INTEGER NOT NULL, trigger_type TEXT NOT NULL, input JSONB NOT NULL DEFAULT '{}', status execution_status NOT NULL DEFAULT 'QUEUED', result JSONB, error TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), started_at TIMESTAMPTZ, finished_at TIMESTAMPTZ, cancelled_at TIMESTAMPTZ
);
CREATE INDEX executions_workflow_created ON executions(workflow_id, created_at DESC);
CREATE TABLE execution_events (id BIGSERIAL PRIMARY KEY, execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE, at TIMESTAMPTZ NOT NULL DEFAULT now(), type TEXT NOT NULL, node_id TEXT, payload JSONB NOT NULL DEFAULT '{}');
CREATE TABLE durable_jobs (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), execution_id UUID NOT NULL REFERENCES executions(id) ON DELETE CASCADE, run_at TIMESTAMPTZ NOT NULL DEFAULT now(), lease_until TIMESTAMPTZ, worker_id TEXT, attempts INT NOT NULL DEFAULT 0, done_at TIMESTAMPTZ, UNIQUE(execution_id));
CREATE INDEX durable_jobs_ready ON durable_jobs(run_at) WHERE done_at IS NULL;

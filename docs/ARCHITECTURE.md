# Architecture

PostgreSQL is the source of truth for workflow drafts, immutable workflow versions, executions, events, and durable jobs. The API only enqueues execution work; workers lease jobs with row locking, preventing normal duplicate claims. A lease expiry permits recovery after a worker dies.

The current worker finalizes a minimal execution record. The next engine increment must load the immutable version, traverse the validated graph, and persist a node-execution/event record for every step.

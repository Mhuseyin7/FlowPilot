# Threat model

Protected assets include workflow definitions, credential material, webhook payloads, and execution history. Principal threats are cross-tenant access, secret leakage, SSRF through outbound connectors, forged/replayed webhooks, queue duplication, and unsafe transformation/code execution.

The current implementation reduces duplicate job claims using durable PostgreSQL leasing. It intentionally contains no outbound connector or arbitrary-code feature. Before adding connectors, outbound IP/DNS/redirect validation and structured redaction are mandatory.

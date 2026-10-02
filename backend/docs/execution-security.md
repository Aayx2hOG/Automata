Apply migrations through `000011_outbox_leases` before starting the server.
Existing workflows receive independent random webhook secrets. Owners can retrieve
`webhook_secret` through authenticated workflow create/get/list responses.

Webhook POST requests now require `X-Webhook-Timestamp` (Unix seconds) and
`X-Webhook-Signature` (hex HMAC-SHA256). Use the workflow's secret string as the
HMAC key and sign the timestamp, a literal period, and the exact request body:
`timestamp + "." + body`. Requests outside a five-minute clock window are rejected.
Signatures may be replayed within that window; consumers requiring exactly-once
business effects should deduplicate event IDs. JSON bodies are limited to 1 MiB.

HTTP request nodes only connect to public HTTP(S) destinations, without URL
credentials or environment proxies. Every resolved address must pass validation;
connections use the validated IP directly, including after redirects. Responses
are limited to 4 MiB after decompression. Internal service URLs no longer work.

Manual, webhook, and scheduled executions share the PostgreSQL outbox consumer.
Run creation and dispatch are atomic; webhook payloads are persisted with jobs.
Retries run under a renewable lease without holding a database transaction.
Shutdown releases unfinished work; process loss makes it available after lease
expiry. Stale owners cannot acknowledge results. Side effects are at-least-once
and may repeat after a crash. The dispatcher and workers start even without schedules.

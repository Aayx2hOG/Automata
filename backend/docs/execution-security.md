# Execution security

Apply all migrations through `000013_node_execution_state` before starting this
version. See [upgrade and execution semantics](scheduled-execution.md) for the
required worker shutdown and compatibility constraints.

## Authentication and ownership

Authenticated workflow, schedule, and run routes enforce workflow ownership.
Passwords use Argon2id. Refresh tokens are hashed in storage and rotate atomically;
revoked or expired tokens and inactive users cannot refresh. Access tokens remain
valid until their configured expiry. Use TLS at the application edge and protect
the JWT signing secret through the deployment platform.

## Webhooks

Each workflow has an independent random secret, available to its owner through
authenticated workflow responses. POST requests require `X-Webhook-Timestamp`
(Unix seconds) and `X-Webhook-Signature` (hex HMAC-SHA256). Sign
`timestamp + "." + exact_request_body` using the secret string directly as the key.
Timestamps must fall within five minutes of server time.

Signatures authenticate a request but permit replay within that window. Separate
deliveries create separate runs. Consumers requiring deduplication must retain
business event IDs. Request bodies are limited to 1 MiB before handler dispatch.

## Outbound HTTP and retries

HTTP nodes connect only to public HTTP(S) destinations, without URL credentials
or environment proxies. Every DNS result must pass address validation; connections
use validated IPs directly, including redirect destinations. Response bodies are
limited to 4 MiB after decompression.

Each durable HTTP node sends a stable engine-owned `Idempotency-Key`, overriding a
configured header of the same name. Completed node outputs are checkpointed under
a fenced execution lease. These controls do not make external effects exactly
once: destinations must honor the key, and an interrupted uncheckpointed request
may already have succeeded. Retries are opt-in per node and default to one attempt.
Review destination idempotency guarantees before enabling retries for writes.

HTTP non-2xx responses remain normal node outputs with `success=false`; branch on
the result as appropriate. They do not automatically produce a retryable error.

## Proxy trust and resource limits

`HTTP_TRUSTED_PROXIES` defaults to empty, so caller-supplied forwarding headers are
ignored. Configure only actual reverse proxy CIDRs. The edge must strip or safely
append forwarding headers. The API scans `X-Forwarded-For` from the connected peer
backwards to the first untrusted hop. Malformed chains fall back to the peer IP.
Do not trust an entire network that also contains untrusted clients.

Auth and webhook IP throttles are per process and track at most 10000 IPs per
limiter. New identities are rejected when that tracking capacity is full until
idle entries expire. Database-backed run quotas apply across all application
instances and all trigger types. See the [configuration table](../../readme.md#configuration).
The test bypass only works with `ENV=test`; never deploy with that environment.

Request bodies are capped at 1 MiB, including trailing data. Graphs are capped at
100 nodes and 500 edges. These limits complement deployment-level connection
limits, TLS, and network access controls.

## Sensitive data

Graph configuration, node outputs, webhook payloads, and webhook secrets are
stored in PostgreSQL without application-level encryption. The former
`CREDENTIALS_ENCRYPTION_KEY` variable did not implement encryption and has been
removed. Restrict database and backup access. Managed credential references,
secret redaction, and data-retention policies are not implemented yet.

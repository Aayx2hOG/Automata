# Database backup and recovery

## What is provided

The Compose files run PostgreSQL with a persistent data volume. They do **not**
configure automatic backups, WAL archiving, pgBackRest, or point-in-time recovery.
Production backup scheduling and storage are deployment responsibilities.

CI performs a logical dump/restore smoke test using PostgreSQL's bundled tools.
This verifies schema/data restoration in an isolated CI database. It does not
verify your production storage, credentials, backup schedule, or disaster recovery.

## Backup procedure

Use PostgreSQL client tools compatible with the server version. Set `DATABASE_URL`
in a secure environment; do not commit credentials. A logical backup can be taken
while the database is running:

```bash
umask 077
pg_dump --format=custom --file=automata.dump "$DATABASE_URL"
pg_restore --list automata.dump
```

Schedule this command using your deployment platform or backup service, check its
exit status, and copy successful backups to encrypted off-host storage. Define
retention, monitor backup age and failures, and regularly restore a downloaded copy.
Protect encryption keys separately from backups. A dump does not back up cluster
roles or deployment secrets; manage those independently.

Your recovery point is the last successful dump. If that is insufficient, configure
managed PostgreSQL point-in-time recovery or an independently operated WAL backup
solution and test its restore procedure. The local Docker volume alone is not a
backup and does not protect against host loss.

## Restore drill and cutover

Restore into a new, empty database using an appropriately privileged recovery role.
Set `RESTORE_DATABASE_URL` to that isolated database, never the live database:

```bash
pg_restore --exit-on-error --no-owner --no-acl \
  --dbname="$RESTORE_DATABASE_URL" automata.dump
psql "$RESTORE_DATABASE_URL" -v ON_ERROR_STOP=1 \
  -c 'SELECT version, dirty FROM schema_migrations;' \
  -c 'SELECT status, count(*) FROM workflow_runs GROUP BY status;' \
  -c 'SELECT count(*) FROM workflow_node_runs;'
```

Verify representative users, workflow versions, run results, and node checkpoints.
Use the application version matching the restored migration level; apply required
upgrades before starting a newer binary. Restore application role grants before
cutover. Record the achieved recovery point and recovery time.

Before switching the live application to the restored database, stop all workers
and review unfinished runs. Restoring a backup can lose recent checkpoints while
external effects still exist. Stable run/node idempotency keys help only when the
external service honors them and still retains their history. Do not automatically
replay uncertain writes. Start workers after reconciliation, verify a controlled
workflow, and confirm the next backup succeeds.

## Database outages

Committed outbox jobs survive application restarts and resume when the database
returns. New triggers are not buffered outside PostgreSQL. Recognized connectivity
and connection-capacity failures return `503` with `Retry-After: 5`. Admission
capacity failures also return `503`, with `Retry-After: 60`.

A lost connection during commit can leave a caller unsure whether a request was
accepted. Retrying that request may create another run; inbound request deduplication
is not implemented. Sign webhook retries with a fresh timestamp once the original
signature expires. Monitor application database errors, disk space, connection
usage, pending-job growth, and backup age through your deployment platform.

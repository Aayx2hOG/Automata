# Database backup and outage recovery

`docker compose up -d --build` now starts PostgreSQL with WAL archiving and a
pgBackRest backup worker. Compose 2.24+ is required. The existing `postgres_data`
volume is retained. Restarting PostgreSQL to enable archiving briefly interrupts
connections; deploy during a maintenance window.

The worker creates a full backup on startup and every 24 hours after a successful
backup, retains seven full backups with their required WAL, and retries failures
every minute. PostgreSQL switches active WAL segments after 60 seconds of activity.
This is asynchronous archival: recent transactions can be lost if the database
disk fails before WAL reaches the repository. Failed archival retains WAL locally
and retries; prolonged repository outages can fill the database disk.

## Off-server storage

The default `postgres_backups` Docker volume is for local development. It survives
database-container loss, but NOT loss of the host or `docker compose down -v`.
Before production, copy `ops/postgres/backup.env.example` to `.env.backup` in the
repository root, fill in a dedicated S3 bucket/prefix, credentials and encryption
passphrase, then recreate both services:

```sh
docker compose up -d --build --force-recreate postgres backup
docker compose exec -u postgres backup pgbackrest --stanza=automata info
docker compose exec -u postgres backup pgbackrest --stanza=automata check
```

Both services read `.env.backup`; Git ignores it. Store the encryption passphrase
and storage credentials separately from the database host. Each independent
cluster needs its own repository prefix. Changing repositories requires a new
successful full backup before the new destination offers recovery coverage.
Use a bucket lifecycle policy that does not delete objects pgBackRest still needs.

## Monitoring

```sh
docker compose ps
docker compose logs backup
docker compose exec -u postgres backup pgbackrest --stanza=automata info
docker compose exec -u postgres postgres psql -U automata -d postgres -c \
  'SELECT archived_count, failed_count, last_archived_time, last_failed_time FROM pg_stat_archiver;'
```

The backup healthcheck fails if the last successful backup is older than 26 hours
or a WAL archive check fails. Connect Docker unhealthy status, backup failures,
and database disk usage to deployment alerts; Docker does not send alerts or
restart unhealthy containers automatically. The first backup may take longer than
the configured ten-minute startup grace period on a large database.

## Restore into a fresh volume

Stop application workers first: restored pending jobs may replay external effects.
Keep the original data volume until recovery has been verified. The restore tool
refuses a nonempty target and does not automatically overwrite the live database.

For S3 recovery on a replacement host, place the same `.env.backup` there, build
the image, and restore into a new volume:

```sh
docker compose build postgres
docker run --rm --env-file .env.backup \
  -v automata_recovered_data:/var/lib/postgresql/data \
  --entrypoint restore-empty automata-postgres:16-backup
```

To stop at a particular time, append
`--type=time --target="2026-09-27 12:00:00+00" --target-action=promote` to the
restore command. The requested time must be covered by the retained backup/WAL
history. Without a time target, PostgreSQL replays all available archived WAL.

For a local repository, omit `--env-file` and additionally mount the original
backup volume with `-v <project>_postgres_backups:/var/lib/pgbackrest`.

Start the restored database in isolation so PostgreSQL performs WAL replay:

```sh
docker run -d --name automata-recovery-check --network none --env-file .env.backup \
  -v automata_recovered_data:/var/lib/postgresql/data automata-postgres:16-backup
docker exec -u postgres automata-recovery-check psql -U automata -d automata -c \
  'SELECT pg_is_in_recovery();'
```

For local recovery, replace the environment-file option with the backup-volume
mount here too. Inspect logs, verify recovered application data and ensure
`pg_is_in_recovery()` is false before cutover. Stop the verification container;
configure Compose's `postgres_data` volume with `external: true` and
`name: automata_recovered_data` in a local override, then start PostgreSQL and
the backup worker with that override. Ensure a new backup succeeds before
resuming application workers. Never start two PostgreSQL servers on one volume.

## Outage contract and restore drill

Manual and webhook triggers return HTTP 503 with `Retry-After: 5` for recognized
database connectivity, shutdown, recovery and connection-capacity failures.
Webhook verification preserves database errors instead of treating them as bad
signatures. Callers should retry with backoff; sign webhook retries with a fresh
timestamp because signatures expire after five minutes.

No new request is durably buffered outside PostgreSQL. Existing committed outbox
jobs resume when the database returns. A connection loss during commit can leave
the caller uncertain whether a request committed; a retry may create another run.
External effects must tolerate duplicates. Missed cron intervals are still skipped.

Run `sh ops/postgres/test-recovery.sh` on a Docker-enabled machine. It uses isolated
volumes, takes a full backup, writes a later row, archives WAL, stops the source,
restores into a fresh volume and asserts both rows survived. It also verifies the
nonempty-target guard. CI runs this drill; repeat it against the configured remote
repository on a separate recovery host before relying on production backups.

Reference: [pgBackRest documentation](https://pgbackrest.org/user-guide.html).

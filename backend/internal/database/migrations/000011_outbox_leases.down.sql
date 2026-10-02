DROP INDEX schedule_outbox_lease_expiry;
ALTER TABLE schedule_outbox
    DROP CONSTRAINT schedule_outbox_lease_pair,
    DROP COLUMN lease_token,
    DROP COLUMN lease_until;

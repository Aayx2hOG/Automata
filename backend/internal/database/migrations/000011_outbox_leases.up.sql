ALTER TABLE schedule_outbox
    ADD COLUMN lease_token UUID,
    ADD COLUMN lease_until TIMESTAMPTZ,
    ADD CONSTRAINT schedule_outbox_lease_pair CHECK (
        (lease_token IS NULL) = (lease_until IS NULL)
    );
CREATE INDEX schedule_outbox_lease_expiry ON schedule_outbox(lease_until);

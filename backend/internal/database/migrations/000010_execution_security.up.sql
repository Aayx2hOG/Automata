ALTER TABLE workflows ADD COLUMN webhook_secret TEXT NOT NULL DEFAULT encode(gen_random_bytes(32), 'hex');
ALTER TABLE schedule_outbox ADD COLUMN seed_data JSONB;

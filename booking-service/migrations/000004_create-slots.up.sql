CREATE TABLE IF NOT EXISTS slot (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' check (status in ('active', 'cancelled')),
    trainer_id UUID NOT NULL REFERENCES trainer(id) ON DELETE CASCADE,
    client_id UUID NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),

    CHECK (start_time < end_time)
);

CREATE UNIQUE INDEX slot_active_slot_unique
    ON slot (trainer_id, start_time)
    WHERE status = 'active';
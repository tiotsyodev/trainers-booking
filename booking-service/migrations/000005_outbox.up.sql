create table if not exists outbox_events (
    id UUID primary key default gen_random_uuid(),
    partition_key text not null,
    payload bytea not null,
    created_at timestamptz not null default now(),
    sent_at timestamptz,
    attempts int not null default 0  
);

create index if not exists idx_outbox_events on outbox_events (created_at) where sent_at is null;
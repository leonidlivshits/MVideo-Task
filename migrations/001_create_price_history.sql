create table if not exists price_history (
    id bigserial primary key,
    good_id bigint not null check (good_id > 0),
    create_at timestamptz not null default now(),
    price integer not null check (price >= 0)
);

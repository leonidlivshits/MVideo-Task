create table if not exists price_history (
    id bigserial primary key,
    good_id bigint not null,
    create_at timestamptz not null default clock_timestamp(),
    price integer not null check (price >= 0)
);

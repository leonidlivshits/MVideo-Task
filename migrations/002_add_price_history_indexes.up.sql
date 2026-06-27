create index price_history_good_id_create_at_id_idx
    on price_history (good_id, create_at desc, id desc)
    include (price);

create index price_history_create_at_good_id_id_idx
    on price_history (create_at, good_id, id)
    include (price);

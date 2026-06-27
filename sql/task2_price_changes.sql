/*
Задание 2.

create table price_history(
    good_id bigint,
    create_at timestamptz,
    price integer
);

изменением цены считается строка, в которой price отличается от предыдущей
цены этого же товара. Первая цена товара изменением не считается, потому что
у нее нет предыдущего значения.
*/

select
    good_id,
    count(*) as changes_count
from (
    select
        good_id,
        price,
        lag(price) over (
            partition by good_id
            order by create_at
        ) as previous_price
    from price_history
) as ordered_prices
where previous_price is not null
    and price <> previous_price
group by good_id
having count(*) > 3
order by good_id;

package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"mvideo-task/internal/domain"
	"mvideo-task/internal/service"
)

const insertPriceSQL = `
	insert into price_history (good_id, price)
	values ($1, $2)
	returning good_id, price, create_at
`

const getPricesAtSQL = `
	select distinct on (good_id)
		good_id, price, create_at
	from price_history
	where good_id = any($1::bigint[])
		and create_at <= $2
	order by good_id, create_at desc, id desc
`

const getHistorySQL = `
	select good_id, price, create_at
	from price_history
	where create_at >= $1
		and create_at < $2
		and (cardinality($3::bigint[]) = 0 or good_id = any($3::bigint[]))
	order by good_id, create_at, id
`

type PriceRepository struct {
	pool *pgxpool.Pool
}

var _ service.PriceRepository = (*PriceRepository)(nil)

func NewPriceRepository(pool *pgxpool.Pool) *PriceRepository {
	return &PriceRepository{pool: pool}
}

func (r *PriceRepository) InsertPrices(ctx context.Context, prices []domain.PricePoint) ([]domain.PricePoint, error) {
	if len(prices) == 0 {
		return []domain.PricePoint{}, nil
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin insert prices transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	created := make([]domain.PricePoint, 0, len(prices))
	for _, price := range prices {
		var (
			goodID   int64
			value    int32
			createAt time.Time
		)

		err := tx.QueryRow(ctx, insertPriceSQL, int64(price.GoodID), int(price.Price)).Scan(&goodID, &value, &createAt)
		if err != nil {
			return nil, fmt.Errorf("insert price for good %d: %w", price.GoodID, err)
		}

		created = append(created, domain.PricePoint{
			GoodID:   domain.GoodID(goodID),
			Price:    domain.Price(value),
			CreateAt: createAt,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit insert prices transaction: %w", err)
	}

	return created, nil
}

func (r *PriceRepository) GetPricesAt(ctx context.Context, goodIDs []domain.GoodID, at time.Time) ([]domain.PriceAt, error) {
	if len(goodIDs) == 0 {
		return []domain.PriceAt{}, nil
	}

	rows, err := r.pool.Query(ctx, getPricesAtSQL, goodIDValues(goodIDs), at)
	if err != nil {
		return nil, fmt.Errorf("query prices at: %w", err)
	}
	defer rows.Close()

	found := make(map[domain.GoodID]domain.PriceAt, len(goodIDs))
	for rows.Next() {
		var (
			goodID   int64
			price    int32
			createAt time.Time
		)

		if err := rows.Scan(&goodID, &price, &createAt); err != nil {
			return nil, fmt.Errorf("scan price at: %w", err)
		}

		value := domain.Price(price)
		timeValue := createAt
		found[domain.GoodID(goodID)] = domain.PriceAt{
			GoodID:   domain.GoodID(goodID),
			Price:    &value,
			CreateAt: &timeValue,
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate prices at: %w", err)
	}

	prices := make([]domain.PriceAt, 0, len(goodIDs))
	for _, goodID := range goodIDs {
		if price, ok := found[goodID]; ok {
			prices = append(prices, price)
			continue
		}

		prices = append(prices, domain.PriceAt{GoodID: goodID})
	}

	return prices, nil
}

func (r *PriceRepository) GetHistory(ctx context.Context, filter service.HistoryFilter) ([]domain.PricePoint, error) {
	rows, err := r.pool.Query(ctx, getHistorySQL, filter.From, filter.To, goodIDValues(filter.GoodIDs))
	if err != nil {
		return nil, fmt.Errorf("query price history: %w", err)
	}
	defer rows.Close()

	history := make([]domain.PricePoint, 0)
	for rows.Next() {
		var (
			goodID   int64
			price    int32
			createAt time.Time
		)

		if err := rows.Scan(&goodID, &price, &createAt); err != nil {
			return nil, fmt.Errorf("scan price history: %w", err)
		}

		history = append(history, domain.PricePoint{
			GoodID:   domain.GoodID(goodID),
			Price:    domain.Price(price),
			CreateAt: createAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate price history: %w", err)
	}

	return history, nil
}

func goodIDValues(goodIDs []domain.GoodID) []int64 {
	values := make([]int64, 0, len(goodIDs))
	for _, goodID := range goodIDs {
		values = append(values, int64(goodID))
	}

	return values
}

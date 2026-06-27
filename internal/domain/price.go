package domain

import "time"

type GoodID int64

type Price int

type PricePoint struct {
	GoodID   GoodID
	Price    Price
	CreateAt time.Time
}

type PriceAt struct {
	GoodID   GoodID
	Price    *Price
	CreateAt *time.Time
}

type Period struct {
	From time.Time
	To   time.Time
}

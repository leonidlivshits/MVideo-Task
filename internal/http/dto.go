package http

import "time"

type SetPricesRequest struct {
	Items []SetPriceItem `json:"items"`
}

type SetPriceItem struct {
	GoodID int64 `json:"good_id"`
	Price  int   `json:"price"`
}

type SetPricesResponse struct {
	Items []SetPriceResponse `json:"items"`
}

type SetPriceResponse struct {
	GoodID   int64     `json:"good_id"`
	Price    int       `json:"price"`
	CreateAt time.Time `json:"create_at"`
}

type PriceResponse struct {
	GoodID   int64      `json:"good_id"`
	Price    *int       `json:"price"`
	CreateAt *time.Time `json:"create_at"`
}

type PricesAtResponse struct {
	At    time.Time       `json:"at"`
	Items []PriceResponse `json:"items"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

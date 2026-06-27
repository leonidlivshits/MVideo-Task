package service

import "errors"

var (
	ErrEmptyPrices   = errors.New("empty prices")
	ErrEmptyGoodIDs  = errors.New("empty good ids")
	ErrInvalidGoodID = errors.New("invalid good id")
	ErrInvalidPrice  = errors.New("invalid price")
	ErrInvalidPeriod = errors.New("invalid period")
)

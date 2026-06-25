package service

import "errors"

var (
	ErrInvalidGoodID = errors.New("invalid good id")
	ErrInvalidPrice  = errors.New("invalid price")
	ErrInvalidPeriod = errors.New("invalid period")
)

package dto

type SymbolStat struct {
	Symbol   string  `json:"symbol"`
	Trades   int64   `json:"trades"`
	Volume   int64   `json:"volume"`
	Turnover float64 `json:"turnover"`
}

type UserRank struct {
	UserID int64 `json:"user_id"`
	Volume int64 `json:"volume"`
}

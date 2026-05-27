package dto

type RecoveryDirtyOrder struct {
	ID             uint64 `json:"id"`
	Symbol         string `json:"symbol"`
	Status         string `json:"status"`
	Quantity       int64  `json:"quantity"`
	FilledQuantity int64  `json:"filled_quantity"`
	Reason         string `json:"reason"`
}

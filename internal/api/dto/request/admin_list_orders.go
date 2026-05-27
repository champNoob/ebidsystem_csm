package request

type AdminListOrdersRequest struct {
	UserID   *int64 `form:"user_id"`
	Symbol   string `form:"symbol"`
	Status   string `form:"status"`
	Side     string `form:"side"`
	Type     string `form:"type"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

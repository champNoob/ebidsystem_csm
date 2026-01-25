package request

type ListOrdersRequest struct {
	Status string `form:"status"` // current | history | all
}

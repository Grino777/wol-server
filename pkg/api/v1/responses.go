package v1

type SuccessfulMessage string

var (
	Success SuccessfulMessage = "success"
	Created SuccessfulMessage = "created"
	Deleted SuccessfulMessage = "deleted"
	Updated SuccessfulMessage = "updated"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SuccessfulResponse struct {
	Status string `json:"status"`
}

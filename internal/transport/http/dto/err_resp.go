package dto

type ErrorResponse struct {
	Err string `json:"error" example:"Internal server error"`
}
package dto

type EmptyResponse struct {
	BaseResponse
}
type BaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

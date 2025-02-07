package controller

type ErrorResponse struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func ErrorResponseFromErr(msg string, err error) *ErrorResponse {
	return &ErrorResponse{msg, err.Error()}
}

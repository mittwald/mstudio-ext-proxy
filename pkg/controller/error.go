package controller

type errorResponse struct {
	Message string `json:"message"`
	Details string `json:"details"`
}

func errorResponseFromErr(msg string, err error) *errorResponse {
	return &errorResponse{msg, err.Error()}
}

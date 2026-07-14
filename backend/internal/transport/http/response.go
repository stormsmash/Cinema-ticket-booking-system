package httptransport

import "github.com/gin-gonic/gin"

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorResponse{
		Error: errorBody{
			Code:    code,
			Message: message,
		},
	})
}

package response

import "github.com/gin-gonic/gin"

type Error struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

type Data struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func Response(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

func MessageResponse(c *gin.Context, statusCode int, message string) {
	Response(c, statusCode, Data{
		Code:    statusCode,
		Message: message,
	})
}

func ErrorResponse(c *gin.Context, statusCode int, message string) {
	Response(c, statusCode, Error{
		Code:  statusCode,
		Error: message,
	})
	c.Abort()
}

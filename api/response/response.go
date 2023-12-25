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

// func RespondWithSuccess(c *gin.Context, data interface{}) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"Success": true,
// 		"Data":    data,
// 	})
// }

// func RespondWithError(c *gin.Context, statusCode int, message string) {
// 	c.JSON(statusCode, gin.H{
// 		"Success": false,
// 		"Error":   message,
// 	})
// 	c.Abort()
// }

// func RespondNotFound(c *gin.Context, message string) {
// 	RespondWithError(c, http.StatusNotFound, message)
// }

// func RespondBadRequest(c *gin.Context, message string) {
// 	RespondWithError(c, http.StatusBadRequest, message)
// }

// func RespondInternalServerError(c *gin.Context, message string) {
// 	RespondWithError(c, http.StatusInternalServerError, message)
// }

// func HandleError(c *gin.Context, err error) {
// 	if err != nil {
// 		RespondInternalServerError(c, "Internal server error")
// 	}
// }

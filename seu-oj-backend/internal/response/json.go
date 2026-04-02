package response

import "github.com/gin-gonic/gin"

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Body{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

func Error(c *gin.Context, message string) {
	c.JSON(200, Body{
		Code:    1,
		Message: message,
		Data:    nil,
	})
}

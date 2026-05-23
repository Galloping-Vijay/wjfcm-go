package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	Data       any    `json:"data,omitempty"`
	Count      int64  `json:"count,omitempty"`
	Meta       any    `json:"meta,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	CreateTime string `json:"create_time"`
}

func OK(c *gin.Context, msg string, data any) {
	c.JSON(http.StatusOK, Body{
		Code:       0,
		Msg:        msg,
		Data:       data,
		RequestID:  requestID(c),
		CreateTime: now(),
	})
}

func Page(c *gin.Context, msg string, data any, count int64) {
	c.JSON(http.StatusOK, Body{
		Code:       0,
		Msg:        msg,
		Data:       data,
		Count:      count,
		RequestID:  requestID(c),
		CreateTime: now(),
	})
}

func Error(c *gin.Context, status int, code int, msg string) {
	c.JSON(status, Body{
		Code:       code,
		Msg:        msg,
		RequestID:  requestID(c),
		CreateTime: now(),
	})
}

func requestID(c *gin.Context) string {
	if c == nil {
		return ""
	}
	return c.GetString("request_id")
}

func now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

package handler

import (
	"errors"
	"net/http"
	"os"

	"wjfcms-go/internal/requestlog"
	"wjfcms-go/internal/response"

	"github.com/gin-gonic/gin"
)

type RequestLogHandler struct{}

func NewRequestLogHandler() *RequestLogHandler {
	return &RequestLogHandler{}
}

func (h *RequestLogHandler) Show(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		requestID = c.Query("request_id")
	}
	entry, path, err := requestlog.Find(requestID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			response.Error(c, http.StatusNotFound, 1, "请求日志不存在")
			return
		}
		response.Error(c, http.StatusNotFound, 1, "请求日志不存在")
		return
	}
	response.OK(c, "获取成功", gin.H{
		"path": path,
		"log":  entry,
	})
}

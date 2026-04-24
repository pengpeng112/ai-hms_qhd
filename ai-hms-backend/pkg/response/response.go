package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ============ 标准响应格式（符合 Notion 规范）============

// SuccessResponse 成功响应
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     ErrorInfo `json:"error"`
	Timestamp string    `json:"timestamp"`
}

// ErrorInfo 错误详情
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// PaginationResponse 分页响应
type PaginationResponse struct {
	Items      interface{}    `json:"items"`
	Pagination PaginationMeta `json:"pagination"`
}

// PaginationMeta 分页元数据
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// ============ 辅助函数 ============

// Success 成功响应（标准格式）
func Success(c *gin.Context, data interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, SuccessResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// SuccessCreated 创建成功响应（标准格式，HTTP 201）
func SuccessCreated(c *gin.Context, data interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusCreated, SuccessResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// Paginated 分页响应（标准格式）
func Paginated(c *gin.Context, items interface{}, page, pageSize int, total int64) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	Success(c, PaginationResponse{
		Items: items,
		Pagination: PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// Error 错误响应（标准格式）
func Error(c *gin.Context, code int, errCode string, message string) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(code, ErrorResponse{
		Success: false,
		Error: ErrorInfo{
			Code:    errCode,
			Message: message,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// ErrorWithDetails 带详情的错误响应
func ErrorWithDetails(c *gin.Context, code int, errCode string, message string, details interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(code, ErrorResponse{
		Success: false,
		Error: ErrorInfo{
			Code:    errCode,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

// ============ HTTP 错误响应便捷方法 ============

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized 401 错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden 403 错误
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound 404 错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// InternalError 500 错误
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

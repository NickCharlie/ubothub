package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ubothub/backend/pkg/errcode"
)

// Response represents the standard API response structure.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PagedData represents paginated data with metadata.
type PagedData struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// OK sends a success response with data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// OKWithMessage sends a success response with a custom message.
func OKWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: message,
	})
}

// OKPaged sends a paginated success response.
func OKPaged(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data: PagedData{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// Error sends an error response based on the provided ErrCode.
func Error(c *gin.Context, err *errcode.ErrCode) {
	c.JSON(err.Status, Response{
		Code:    err.Code,
		Message: err.Message,
	})
}

// ErrorWithMessage sends an error response with a custom message.
func ErrorWithMessage(c *gin.Context, err *errcode.ErrCode, message string) {
	c.JSON(err.Status, Response{
		Code:    err.Code,
		Message: message,
	})
}

// ValidationError sends a validation error response with field details.
func ValidationError(c *gin.Context, details interface{}) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Code:    errcode.ErrValidation.Code,
		Message: errcode.ErrValidation.Message,
		Data:    details,
	})
}

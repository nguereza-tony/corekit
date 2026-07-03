package helper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nguereza-tony/corekit/common"
)

const RequestIDHeader = "X-Request-ID"

// GetRequestID extracts or generates a request ID from the Gin context
func GetRequestID(c *gin.Context) string {
	requestID := c.GetHeader(RequestIDHeader)
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return requestID
}

// SetRequestID sets a request ID in the Gin context
func SetRequestID(c *gin.Context) {
	requestID := GetRequestID(c)
	c.Set("request_id", requestID)
	c.Header(RequestIDHeader, requestID)
}

// SetCookie sets an HTTP cookie with SameSite support
func SetCookie(c *gin.Context, name, value string, maxAge int, path, domain string, secure, httpOnly bool, sameSite http.SameSite) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: sameSite,
	})
}

// ClearCookie removes an HTTP cookie
func ClearCookie(c *gin.Context, name, path, domain string, secure, httpOnly bool) {
	SetCookie(c, name, "", -1, path, domain, secure, httpOnly, http.SameSiteLaxMode)
}

func ObjectToJson(obj any) (map[string]any, error) {
	result := map[string]any{}
	data, err := json.Marshal(obj)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// ParseSameSite converts string to http.SameSite
func ParseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(total int64, page, limit int) common.PaginationMeta {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	meta := common.PaginationMeta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	if page < totalPages {
		nextPage := page + 1
		meta.NextPage = &nextPage
	}
	if page > 1 {
		prevPage := page - 1
		meta.PrevPage = &prevPage
	}

	return meta
}

// GetOffset calculates offset for database query
func GetOffset(page, limit int) int {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	return (page - 1) * limit
}

// GetPage returns normalized page value
func GetPage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

// replaceVariables replaces {{VariableName}} placeholders in text with values from data map
func ReplaceTemplateVariables(text string, data map[string]string) string {
	result := text
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

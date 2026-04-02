package v1

import (
	"errors"
	"strconv"
	"strings"

	"github.com/elliotxin/ai-hms-backend/internal/middleware"
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type LogHandler struct {
	service *services.LogService
}

func NewLogHandler() *LogHandler {
	return &LogHandler{
		service: services.NewLogService(),
	}
}

// GetLogs 读取系统日志
// GET /api/v1/settings/logs
func (h *LogHandler) GetLogs(c *gin.Context) {
	query, err := parseLogQuery(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, getErr := h.service.GetLogs(query)
	if getErr != nil {
		response.InternalError(c, getErr.Error())
		return
	}
	response.Success(c, result)
}

func parseLogQuery(c *gin.Context) (services.LogQuery, error) {
	source := services.LogSource(strings.TrimSpace(c.DefaultQuery("source", string(services.LogSourceAll))))
	lines := 0
	level := services.LogLevel(strings.ToUpper(strings.TrimSpace(c.Query("level"))))
	keyword := strings.TrimSpace(c.Query("keyword"))

	if source != services.LogSourceApp && source != services.LogSourceError && source != services.LogSourceAll {
		return services.LogQuery{}, errors.New("invalid source, expected app|error|all")
	}

	if rawLines := strings.TrimSpace(c.Query("lines")); rawLines != "" {
		parsed, err := strconv.Atoi(rawLines)
		if err != nil {
			return services.LogQuery{}, errors.New("invalid lines, expected integer")
		}
		if parsed < 50 || parsed > 1000 {
			return services.LogQuery{}, errors.New("invalid lines, expected 50-1000")
		}
		lines = parsed
	}

	if level != "" && level != services.LogLevelInfo && level != services.LogLevelWarn && level != services.LogLevelError {
		return services.LogQuery{}, errors.New("invalid level, expected INFO|WARN|ERROR")
	}

	return services.LogQuery{
		Source:  source,
		Lines:   lines,
		Keyword: keyword,
		Level:   level,
	}, nil
}

func RegisterLogRoutes(r *gin.RouterGroup) {
	handler := NewLogHandler()
	settings := r.Group("/settings")
	settings.Use(middleware.RequireRoles("ADMIN"))
	{
		settings.GET("/logs", handler.GetLogs)
	}
}

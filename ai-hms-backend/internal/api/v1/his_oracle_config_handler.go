package v1

import (
	"context"
	"strings"
	"time"

	"github.com/elliotxin/ai-hms-backend/internal/integrations/his_oracle"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type HisOracleConfigHandler struct{}

func NewHisOracleConfigHandler() *HisOracleConfigHandler {
	return &HisOracleConfigHandler{}
}

type HisOracleTestRequest struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	Service  string `json:"service" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *HisOracleConfigHandler) TestConnection(c *gin.Context) {
	var req HisOracleTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写完整的连接信息")
		return
	}

	cfg := his_oracle.Config{
		Host:     strings.TrimSpace(req.Host),
		Port:     req.Port,
		Service:  strings.TrimSpace(req.Service),
		Username: strings.TrimSpace(req.Username),
		Password: req.Password,
	}

	client, err := his_oracle.NewClient(cfg)
	if err != nil {
		response.BadRequest(c, "创建连接失败："+err.Error())
		return
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	latency, err := client.TestConnection(ctx)
	if err != nil {
		response.Success(c, gin.H{"connected": false, "error": err.Error()})
		return
	}
	response.Success(c, gin.H{"connected": true, "latency_ms": latency.Milliseconds()})
}

func RegisterHisOracleConfigRoutes(r *gin.RouterGroup) {
	handler := NewHisOracleConfigHandler()
	r.Group("/sync").POST("/his-oracle/test", handler.TestConnection)
}

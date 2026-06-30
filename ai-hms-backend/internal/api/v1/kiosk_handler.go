package v1

import (
	"github.com/elliotxin/ai-hms-backend/internal/services"
	"github.com/elliotxin/ai-hms-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// RegisterKioskRoutes 注册自助站接口路由。需传入 config.KioskConfig 的 token，网关负责校验。
func RegisterKioskRoutes(r *gin.RouterGroup, kioskToken string) {
	svc := services.NewKioskService()

	kiosk := r.Group("/kiosk")
	kiosk.Use(kioskTokenMiddleware(kioskToken))
	{
		kiosk.GET("/health", func(c *gin.Context) {
			response.Success(c, svc.Health())
		})
		kiosk.GET("/patients/lookup", handleKioskLookup(svc))
		kiosk.POST("/pre-signs", handleKioskPreSigns(svc))
		kiosk.POST("/check-in", handleKioskCheckIn(svc))
	}
}

func kioskTokenMiddleware(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Kiosk-Token")
		if token == "" {
			token = c.Query("token")
		}
		if token != expected || expected == "" {
			response.Error(c, 401, "UNAUTHORIZED", "Invalid or missing kiosk token")
			c.Abort()
			return
		}
		c.Next()
	}
}

func handleKioskPreSigns(svc *services.KioskService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.KioskPreSignsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "请求体解析失败: "+err.Error())
			return
		}
		if err := svc.SavePreSigns(req); err != nil {
			response.InternalError(c, "保存体征失败: "+err.Error())
			return
		}
		response.Success(c, gin.H{"saved": true})
	}
}

func handleKioskCheckIn(svc *services.KioskService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req services.KioskCheckInRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "请求体解析失败: "+err.Error())
			return
		}
		if err := svc.CheckIn(req); err != nil {
			response.InternalError(c, "签到失败: "+err.Error())
			return
		}
		response.Success(c, gin.H{"checkedIn": true})
	}
}

func handleKioskLookup(svc *services.KioskService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var q services.KioskLookupQuery
		if err := c.ShouldBindQuery(&q); err != nil {
			response.BadRequest(c, "查询参数解析失败: "+err.Error())
			return
		}
		result, err := svc.LookupPatient(q)
		if err != nil {
			response.InternalError(c, "查患者失败: "+err.Error())
			return
		}
		response.Success(c, result)
	}
}

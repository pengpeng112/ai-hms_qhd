package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func doReq(t *testing.T, setup func(r *gin.Engine)) (*httptest.ResponseRecorder, map[string]interface{}) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	setup(r)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/t", nil)
	r.ServeHTTP(w, req)
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("响应不是合法 JSON: %v (body=%s)", err, w.Body.String())
	}
	return w, body
}

// 2xx 裸 JSON → {success:true,data,timestamp}
func TestWrapperWrapsSuccess(t *testing.T) {
	w, body := doReq(t, func(r *gin.Engine) {
		r.Use(responseWrapper())
		r.GET("/t", func(c *gin.Context) { c.JSON(200, gin.H{"hello": "world"}) })
	})
	if w.Code != 200 || body["success"] != true {
		t.Fatalf("期望 200 + success=true, got %d %v", w.Code, body)
	}
	data, _ := body["data"].(map[string]interface{})
	if data["hello"] != "world" {
		t.Fatalf("data 未透传: %v", body)
	}
}

// 4xx 历史格式 {code:403,error:"..."} → 归一为 {success:false,error:{code,message}}
func TestWrapperNormalizesLegacyError(t *testing.T) {
	w, body := doReq(t, func(r *gin.Engine) {
		r.Use(responseWrapper())
		r.GET("/t", func(c *gin.Context) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "error": "缺少有效租户"})
		})
	})
	if w.Code != 403 || body["success"] != false {
		t.Fatalf("期望 403 + success=false, got %d %v", w.Code, body)
	}
	e, _ := body["error"].(map[string]interface{})
	if e["code"] != "403" || e["message"] != "缺少有效租户" {
		t.Fatalf("错误体未归一: %v", body)
	}
}

// 4xx 仅 {error:"..."} 无 code → code 回退 HTTP 状态码
func TestWrapperNormalizesErrorWithoutCode(t *testing.T) {
	_, body := doReq(t, func(r *gin.Engine) {
		r.Use(responseWrapper())
		r.GET("/t", func(c *gin.Context) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的排班ID"})
		})
	})
	e, _ := body["error"].(map[string]interface{})
	if e["code"] != "400" || e["message"] != "无效的排班ID" {
		t.Fatalf("缺省 code 未回退状态码: %v", body)
	}
}

// 已含 success 字段 → 原样透传，不二次包装
func TestWrapperPassesThroughUnified(t *testing.T) {
	_, body := doReq(t, func(r *gin.Engine) {
		r.Use(responseWrapper())
		r.GET("/t", func(c *gin.Context) {
			c.JSON(200, gin.H{"success": true, "data": gin.H{"x": 1}})
		})
	})
	if _, nested := body["data"].(map[string]interface{})["success"]; nested {
		t.Fatalf("发生二次包装: %v", body)
	}
}

// 关键回归：wrapper 先注册时，后续中间件 AbortWithStatusJSON 的错误也被归一
// （对应 Register 中 responseWrapper 必须先于 tenantMiddleware 的顺序约束）。
func TestWrapperCatchesMiddlewareAbort(t *testing.T) {
	w, body := doReq(t, func(r *gin.Engine) {
		r.Use(responseWrapper())
		r.Use(func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "error": "未鉴权:请提供 X-Role"})
		})
		r.GET("/t", func(c *gin.Context) { c.JSON(200, gin.H{"never": true}) })
	})
	if w.Code != http.StatusUnauthorized || body["success"] != false {
		t.Fatalf("中间件 abort 未被归一: %d %v", w.Code, body)
	}
	e, _ := body["error"].(map[string]interface{})
	if e["message"] != "未鉴权:请提供 X-Role" {
		t.Fatalf("message 丢失: %v", body)
	}
}

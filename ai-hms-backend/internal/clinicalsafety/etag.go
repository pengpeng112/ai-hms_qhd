package clinicalsafety

import (
	"strconv"
	"strings"
	"time"
)

// 版本令牌走 HTTP ETag / If-Match，使 JSON 响应体契约保持不变（详见设计文档 §2.2）。
// 统一使用弱 ETag：
//   - 整型版本： W/"v<version>"        例如 W/"v7"
//   - 时间戳：   W/"t<unixNano>"        例如 W/"t1718000000000000000"

// VersionETag 将整型版本号编码为弱 ETag。
func VersionETag(version int) string {
	return `W/"v` + strconv.Itoa(version) + `"`
}

// ParseVersionETag 从 ETag / If-Match 头解析整型版本号。ok=false 表示格式不符。
func ParseVersionETag(etag string) (int, bool) {
	body, ok := unwrapWeakETag(etag)
	if !ok || !strings.HasPrefix(body, "v") {
		return 0, false
	}
	n, err := strconv.Atoi(body[1:])
	if err != nil {
		return 0, false
	}
	return n, true
}

// TimeETag 将时间戳编码为弱 ETag（unix 纳秒）。
func TimeETag(t time.Time) string {
	return `W/"t` + strconv.FormatInt(t.UnixNano(), 10) + `"`
}

// ParseTimeETag 从 ETag / If-Match 头解析时间戳（按 unix 纳秒还原为 UTC time.Time）。
func ParseTimeETag(etag string) (time.Time, bool) {
	body, ok := unwrapWeakETag(etag)
	if !ok || !strings.HasPrefix(body, "t") {
		return time.Time{}, false
	}
	n, err := strconv.ParseInt(body[1:], 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	return time.Unix(0, n).UTC(), true
}

// unwrapWeakETag 去掉弱标记与引号，返回内层令牌体。兼容强 ETag 写法。
func unwrapWeakETag(etag string) (string, bool) {
	s := strings.TrimSpace(etag)
	s = strings.TrimPrefix(s, "W/")
	s = strings.TrimPrefix(s, "w/")
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", false
	}
	return s[1 : len(s)-1], true
}

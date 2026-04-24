package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

// LegacyID 用于兼容“数据库 bigint + API 字符串 ID”场景。
type LegacyID int64

// Int64 返回原始 int64 值。
func (l LegacyID) Int64() int64 {
	return int64(l)
}

// MarshalJSON 将 ID 以字符串形式输出，避免前端 number 精度问题。
func (l LegacyID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatInt(int64(l), 10) + `"`), nil
}

// UnmarshalJSON 支持 string / number 双输入。
func (l *LegacyID) UnmarshalJSON(b []byte) error {
	data := bytes.TrimSpace(b)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*l = 0
		return nil
	}

	// string 形态
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("parse LegacyID string failed: %w", err)
		}
		if s == "" {
			*l = 0
			return nil
		}
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid LegacyID string %q: %w", s, err)
		}
		*l = LegacyID(v)
		return nil
	}

	// number 形态
	v, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid LegacyID number %q: %w", string(data), err)
	}
	*l = LegacyID(v)
	return nil
}

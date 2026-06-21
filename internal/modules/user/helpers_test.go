package user

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

// decodeBase64JSON 解析 base64url 编码的 JSON 对象, 测试 helper。
func decodeBase64JSON(t *testing.T, s string) map[string]any {
	t.Helper()
	// JWT 用 base64url (无 padding)
	if pad := len(s) % 4; pad != 0 {
		s += "===="[:4-pad]
	}
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("base64 decode: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("json decode: %v, raw=%s", err, raw)
	}
	return out
}

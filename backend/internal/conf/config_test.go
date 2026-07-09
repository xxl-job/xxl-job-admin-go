package conf

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidateJWTSecret 覆盖空、过短、占位符（应拒）与合法强密钥（应通过）。
func TestValidateJWTSecret(t *testing.T) {
	cases := []struct {
		secret  string
		wantErr bool
	}{
		{"", true},
		{"short", true},
		{"change-me-in-production", true},
		{"CHANGE_ME", true},
		{"this-is-a-strong-secret-32bytes", false},
	}
	for _, c := range cases {
		err := validateJWTSecret(c.secret)
		if c.wantErr && err == nil {
			t.Errorf("secret=%q 期望报错，实际通过", c.secret)
		}
		if !c.wantErr && err != nil {
			t.Errorf("secret=%q 期望通过，实际报错: %v", c.secret, err)
		}
	}
}

// TestLoad_RejectWeakSecret 端到端验证：占位符密钥应使 Load 失败。
func TestLoad_RejectWeakSecret(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte("server:\n  http_addr: \":8080\"\ndb:\n  dsn: \"\"\njwt:\n  secret: \"change-me-in-production\"\n  ttl: \"24h\"\n")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("占位符密钥应被 Load 拒绝，实际成功")
	}
}

package auth

import (
	"testing"
	"time"
)

// TestJWT_SignParse 签发与回解析 claims 一致。
func TestJWT_SignParse(t *testing.T) {
	m := NewJWTManager("test-secret-0123456789", time.Hour)
	tok, err := m.Sign(1, "admin", RoleAdmin)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if tok == "" {
		t.Fatal("空 token")
	}
	c, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if c.UserID != 1 || c.Username != "admin" || c.Role != RoleAdmin {
		t.Errorf("claims 不匹配: %+v", c)
	}
}

// TestJWT_Expired 过期 token 解析失败。
func TestJWT_Expired(t *testing.T) {
	m := NewJWTManager("any-secret-00000000", -time.Second) // 已过期 ttl
	tok, _ := m.Sign(1, "u", RoleUser)
	if _, err := m.Parse(tok); err == nil {
		t.Fatal("过期 token 应解析失败")
	}
}

// TestJWT_Tampered 篡改签名段应失败。
func TestJWT_Tampered(t *testing.T) {
	m := NewJWTManager("secret-test-abcdef", time.Hour)
	tok, _ := m.Sign(1, "u", RoleUser)
	if _, err := m.Parse(tok + "x"); err == nil {
		t.Fatal("篡改 token 应解析失败")
	}
}

// TestJWT_WrongSecret 不同密钥签发的 token 校验失败。
func TestJWT_WrongSecret(t *testing.T) {
	m1 := NewJWTManager("secret-A-0000000000", time.Hour)
	m2 := NewJWTManager("secret-B-0000000000", time.Hour)
	tok, _ := m1.Sign(1, "u", RoleUser)
	if _, err := m2.Parse(tok); err == nil {
		t.Fatal("错密钥应解析失败")
	}
}

// TestJWT_TTL 返回配置的 TTL。
func TestJWT_TTL(t *testing.T) {
	m := NewJWTManager("s", 2*time.Hour)
	if m.TTL() != 2*time.Hour {
		t.Errorf("TTL 不匹配: %v", m.TTL())
	}
}

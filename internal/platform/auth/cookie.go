package auth

import (
	"encoding/json"
	"net/http"
	"olixops/internal/config"
	"time"

	"github.com/gin-gonic/gin"
)

// DefaultConfig 返回默认配置
func DefaultConfig() *config.CookieConfig {
	return &config.CookieConfig{
		Path:     "/",
		MaxAge:   0, // 会话级
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

// Manager Cookie 管理器
type CookieManager struct {
	config *config.CookieConfig
}

// NewManager 创建 Cookie 管理器
func NewCookieManager(cfg *config.CookieConfig) *CookieManager {
	// default config
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return &CookieManager{
		config: cfg,
	}
}

// Set 设置 Cookie
func (m *CookieManager) Set(c *gin.Context, name, value string, maxAge ...int) {

	age := m.config.MaxAge
	if len(maxAge) > 0 && maxAge[0] > 0 {
		age = maxAge[0]
	}

	// 顺序很重要：先设置 SameSite，再 SetCookie
	c.SetSameSite(m.config.SameSite)

	c.SetCookie(
		name,
		value,
		age,
		m.config.Path,
		m.config.Domain,
		m.config.Secure,
		m.config.HttpOnly,
	)
}

// SetWithExpires 设置带过期时间的 Cookie
func (m *CookieManager) SetWithExpires(c *gin.Context, name, value string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = -1 // 立即过期
	}

	c.SetCookie(
		name,
		value,
		maxAge,
		m.config.Path,
		m.config.Domain,
		m.config.Secure,
		m.config.HttpOnly,
	)
	c.SetSameSite(m.config.SameSite)
}

// SetToken 设置 Token Cookie（专门为 JWT 优化）
func (m *CookieManager) SetToken(c *gin.Context, name, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = -1
	}

	c.SetCookie(
		name,
		token,
		maxAge,
		m.config.Path,
		m.config.Domain,
		m.config.Secure,
		m.config.HttpOnly,
	)
	c.SetSameSite(m.config.SameSite)
}

// Get 获取 Cookie
func (m *CookieManager) Get(c *gin.Context, name string) (string, error) {
	return c.Cookie(name)
}

// GetString 获取 Cookie（带默认值）
func (m *CookieManager) GetString(c *gin.Context, name, defaultValue string) string {
	value, err := c.Cookie(name)
	if err != nil {
		return defaultValue
	}
	return value
}

// Clear 清除 Cookie
func (m *CookieManager) Clear(c *gin.Context, name string) {
	c.SetCookie(
		name,
		"",
		-1, // 立即过期
		m.config.Path,
		m.config.Domain,
		m.config.Secure,
		m.config.HttpOnly,
	)
	c.SetSameSite(m.config.SameSite)
}

// ClearMultiple 清除多个 Cookie
func (m *CookieManager) ClearMultiple(c *gin.Context, names ...string) {
	for _, name := range names {
		m.Clear(c, name)
	}
}

// Exists 检查 Cookie 是否存在
func (m *CookieManager) Exists(c *gin.Context, name string) bool {
	_, err := c.Cookie(name)
	return err == nil
}

// IsSecure 检查是否启用安全模式
func (m *CookieManager) IsSecure() bool {
	return m.config.Secure
}

// SetClaims 存储 Claims 到 Cookie（JSON 序列化）
func (m *CookieManager) SetClaims(c *gin.Context, claims interface{}, maxAge ...int) {
	// 序列化 claims
	data, err := json.Marshal(claims)
	if err != nil {
		// 如果序列化失败，记录错误但不中断
		return
	}

	// 存储为字符串
	m.Set(c, "user_claims", string(data), maxAge...)
}

// GetClaims 从 Cookie 获取 Claims
func (m *CookieManager) GetClaims(c *gin.Context, target interface{}) error {
	// 获取 Cookie 值
	value, err := c.Cookie("user_claims")
	if err != nil {
		return err
	}

	// 反序列化
	return json.Unmarshal([]byte(value), target)
}

// SetTokenWithClaims 设置 Token 和 Claims 一起
func (m *CookieManager) SetTokenWithClaims(c *gin.Context, token string, claims interface{}, expiresAt time.Time) {
	// 1. 设置 Token
	m.SetToken(c, "token", token, expiresAt)

	// 2. 设置 Claims
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = -1
	}
	m.SetClaims(c, claims, maxAge)
}

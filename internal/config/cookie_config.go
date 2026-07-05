package config

import "net/http"

// Config Cookie 配置
type CookieConfig struct {
	Path     string        // Cookie 路径，默认 "/"
	Domain   string        // Cookie 域名
	MaxAge   int           // 过期时间（秒），默认 0 表示会话级
	Secure   bool          // 是否只在 HTTPS 下传输
	HttpOnly bool          // 是否禁止 JavaScript 访问
	SameSite http.SameSite // SameSite 策略
}

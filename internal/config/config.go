// Package config 负责加载应用配置。
//
// 加载顺序: 默认值 -> 配置文件(YAML) -> 环境变量。
// 环境变量优先级最高,前缀 OLIXOPS_,层级用下划线分隔。
// 例如 OLIXOPS_DATABASE_HOST 会覆盖 database.host。
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 是应用根配置。
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Log      LogConfig      `mapstructure:"log"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Storage  StorageConfig  `mapstructure:"storage"`
	EnvVault EnvVaultConfig `mapstructure:"env_vault"`
}

// AppConfig 描述应用元信息。
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Env     string `mapstructure:"env"` // dev / test / sim / prod
	Version string `mapstructure:"version"`
}

// ServerConfig 描述 HTTP 服务监听参数。
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Mode            string        `mapstructure:"mode"` // debug / release / test
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// Addr 返回 host:port 形式的监听地址。
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// LogConfig 控制日志输出。
type LogConfig struct {
	Level      string `mapstructure:"level"`  // debug / info / warn / error
	Format     string `mapstructure:"format"` // json / console
	Output     string `mapstructure:"output"` // stdout / stderr / file
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// DatabaseConfig 描述 PostgreSQL 连接参数。
type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"` // 当前仅支持 postgres
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	TimeZone        string        `mapstructure:"timezone"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
	AutoMigrate     bool          `mapstructure:"auto_migrate"`
	LogLevel        string        `mapstructure:"log_level"` // silent / error / warn / info
}

// DSN 生成 PostgreSQL DSN。
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode, d.TimeZone,
	)
}

// RedisConfig 描述 Redis 连接参数。
type RedisConfig struct {
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// AuthConfig 描述认证与 Token 配置。
type AuthConfig struct {
	JWTSecret          string        `mapstructure:"jwt_secret"`
	AccessTokenTTL     time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL    time.Duration `mapstructure:"refresh_token_ttl"`
	Issuer             string        `mapstructure:"issuer"`
	OAuth2Provider     string        `mapstructure:"oauth2_provider"` // keycloak / dex / github 等
	OAuth2ClientID     string        `mapstructure:"oauth2_client_id"`
	OAuth2ClientSecret string        `mapstructure:"oauth2_client_secret"`
	OAuth2DiscoveryURL string        `mapstructure:"oauth2_discovery_url"`
	OAuth2RedirectURL  string        `mapstructure:"oauth2_redirect_url"`
	OAuth2Scopes       []string      `mapstructure:"oauth2_scopes"`
	AllowPasswordLogin bool          `mapstructure:"allow_password_login"`
	PasswordMinLength  int           `mapstructure:"password_min_length"`
}

// StorageConfig 描述文件存储后端。
type StorageConfig struct {
	Driver   string `mapstructure:"driver"` // local / s3 / minio
	LocalDir string `mapstructure:"local_dir"`
	Endpoint string `mapstructure:"endpoint"`
	Bucket   string `mapstructure:"bucket"`
	Region   string `mapstructure:"region"`
	AKID     string `mapstructure:"access_key_id"`
	Secret   string `mapstructure:"secret_access_key"`
	UseSSL   bool   `mapstructure:"use_ssl"`
}

// IsProduction 判断是否为生产环境。
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Env, "prod")
}

// IsDevelopment 判断是否为开发环境。
func (c *Config) IsDevelopment() bool {
	return strings.EqualFold(c.App.Env, "dev")
}

// Load 从 path 指定的目录加载 config.yaml,并合并环境变量。
// path 为空时,默认从 ./configs 加载。
func Load(path string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if path == "" {
		path = "./configs"
	}
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(path)
	v.AddConfigPath(".")

	v.SetEnvPrefix("OLIXOPS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// 配置文件可缺省,环境变量也能完成配置。
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "olixops")
	v.SetDefault("app.env", "dev")
	v.SetDefault("app.version", "0.0.1")

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 15*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)
	v.SetDefault("server.shutdown_timeout", 10*time.Second)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 7)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.compress", true)

	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "olixops")
	v.SetDefault("database.password", "olixops")
	v.SetDefault("database.database", "olixops")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.timezone", "Asia/Shanghai")
	v.SetDefault("database.max_open_conns", 50)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", time.Hour)
	v.SetDefault("database.conn_max_idle_time", 30*time.Minute)
	v.SetDefault("database.auto_migrate", false)
	v.SetDefault("database.log_level", "warn")

	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 32)
	v.SetDefault("redis.min_idle_conns", 4)
	v.SetDefault("redis.dial_timeout", 5*time.Second)
	v.SetDefault("redis.read_timeout", 3*time.Second)
	v.SetDefault("redis.write_timeout", 3*time.Second)

	v.SetDefault("auth.access_token_ttl", 2*time.Hour)
	v.SetDefault("auth.refresh_token_ttl", 720*time.Hour)
	v.SetDefault("auth.issuer", "olixops")
	v.SetDefault("auth.allow_password_login", true)
	v.SetDefault("auth.password_min_length", 8)
	v.SetDefault("auth.oauth2_scopes", []string{"openid", "profile", "email"})

	v.SetDefault("storage.driver", "local")
	v.SetDefault("storage.local_dir", "./data/files")
	v.SetDefault("storage.use_ssl", true)

	v.SetDefault("env_vault.url", "http://localhost:8080")
	v.SetDefault("env_valult.timeout", 30*time.Second)
}

func validate(c *Config) error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server.port: %d", c.Server.Port)
	}
	if c.Database.Host == "" || c.Database.Database == "" {
		return fmt.Errorf("database.host and database.database are required")
	}
	if c.IsProduction() && c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret must be set in production")
	}
	return nil
}

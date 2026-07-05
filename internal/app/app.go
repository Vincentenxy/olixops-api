// Package app 负责应用启动流程与依赖装配。
//
// 装配顺序: config → logger → DB → platform services (auth/audit) → modules → router → http server。
// 生命周期由 Run() 统一管控,支持 signal 优雅关闭。
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"olixops/initialize"
	"olixops/internal/modules/cluster"
	"olixops/internal/platform/audit"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"olixops/internal/config"
	"olixops/internal/interfaces/http/router"
	"olixops/internal/modules/health"
	"olixops/internal/modules/user"
	"olixops/internal/platform/auth"
	"olixops/internal/platform/database"
	"olixops/internal/platform/logger"
)

// Run 启动 HTTP 服务,阻塞直到收到 SIGINT/SIGTERM 后优雅关闭。
func Run() error {
	// 1. init config
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 2. init logger
	lg, err := logger.New(cfg.Log, cfg.App.Name, cfg.App.Env)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer logger.Sync()

	lg.Info("olixops starting",
		zap.String("name", cfg.App.Name),
		zap.String("env", cfg.App.Env),
		zap.String("version", cfg.App.Version),
		zap.String("addr", cfg.Server.Addr()),
	)

	// 3. init db
	db, err := database.New(cfg.Database, lg)
	if err != nil {
		lg.Fatal("init database failed", zap.Error(err))
	}
	if cfg.Database.AutoMigrate {
		if err := db.AutoMigrate(&user.User{}); err != nil {
			return fmt.Errorf("automigrate user: %w", err)
		}
		lg.Info("auto migrate user done")
	}

	// 4. env vault
	ctx := context.Background()
	// envvault.NewLoader(ctx, cfg.EnvVault)

	// 5. initialize
	initialize.InitValidator()

	// 4. recoder
	// TODO platform/audit 下面

	// 3.5) 加密服务 (DEK 来自 envVault 或环境变量, 用于加密 kubeconfig / 敏感字段).
	//
	// 当前 cryptox.New / envvault.NewLoader / envvault.LoadKey 都还是 TODO,
	// 这一段先调本地 initCrypto 占位, 等子模块实现好之后替换内部逻辑.
	//cryptoSvc, err := initCrypto(ctx, cfg.EnvVault, lg)
	//if err != nil {
	//	return fmt.Errorf("init crypto: %w", err)
	//}
	//_ = cryptoSvc // 暂时 unused, 后续 cluster service 注入

	// 相关平台服务
	issuer := auth.NewJWTIssuer(&cfg.Auth)
	recorder := audit.NewMemRecorder(1000, lg)
	cookieManager := auth.NewCookieManager(&cfg.CookieConfig)

	// user module
	userModule, err := user.Assemble(db, recorder, issuer, cookieManager)
	if err != nil {
		lg.Fatal("assemble user module failed", zap.Error(err))
	}

	// cluster module
	clusterModule, err := cluster.Assemble(db, recorder, &cfg.K8sConfig)
	if err != nil {
		lg.Fatal("assemble cluster module failed", zap.Error(err))
	}

	// 7. routers
	modules := []router.ModuleRoutes{
		health.New(),
		clusterModule,
		userModule,
	}
	r := router.New(modules, issuer, cookieManager)

	// 8) http server
	srv := &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 9) 监听信号
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 10) 异步启动,主协程阻塞等信号
	serverErr := make(chan error, 1)
	go func() {
		lg.Info("http server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			lg.Error("http server failed", zap.Error(err))
			return err
		}
	case <-ctx.Done():
		lg.Info("shutdown signal received, draining connections",
			zap.Duration("timeout", cfg.Server.ShutdownTimeout),
		)
	}

	// 11) 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		lg.Error("graceful shutdown failed", zap.Error(err))
		return err
	}

	lg.Info("olixops stopped cleanly")
	_ = time.Now()
	return nil
}

func isSeedEnv(env string) bool {
	e := strings.ToLower(env)
	return e == "dev" || e == "test" || e == ""
}

// seedDevUser 直接通过 Repository 创建 seed 用户, 跳过 service.Create 的复杂校验。
// 已存在则 no-op; hash 用真 bcrypt, 与生产路径一致。
func seedDevUser(ctx context.Context, repo user.Repository, hasher auth.PasswordHasher, username, password string) error {
	_, err := repo.FindByUsername(ctx, username)
	if err == nil {
		return nil // 已存在
	}
	hash, err := hasher.Hash(password)
	if err != nil {
		return err
	}
	u := &user.User{
		ID:           uuid.NewString(),
		Username:     username,
		Email:        username + "@local.dev",
		DisplayName:  username,
		PasswordHash: hash,
		Status:       user.StatusActive,
		Source:       "local",
	}
	return repo.Create(ctx, u)
}

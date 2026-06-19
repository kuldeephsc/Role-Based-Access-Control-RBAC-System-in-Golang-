package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"rbac-platform/internal/analytics"
	"rbac-platform/internal/audit"
	"rbac-platform/internal/auth"
	"rbac-platform/internal/config"
	"rbac-platform/internal/middleware"
	"rbac-platform/internal/notification"
	"rbac-platform/internal/outbox"
	"rbac-platform/internal/platform/cache"
	jwtpkg "rbac-platform/internal/platform/jwt"
	"rbac-platform/internal/platform/logger"
	"rbac-platform/internal/platform/rabbitmq"
	"rbac-platform/internal/platform/redisclient"
	"rbac-platform/internal/platform/tracing"
	"rbac-platform/internal/rbac"
	pgrepo "rbac-platform/internal/repository/postgres"
	"rbac-platform/internal/user"

	"time"
)

func main() {
	cfg := config.Load()
	lg := logger.New()

	// --- Postgres ---
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("postgres connect: %v", err)
	}

	// --- Redis (optional -- nil-safe wrappers degrade gracefully) ---
	rdb := redisclient.New(cfg.RedisAddr, cfg.RedisPassword, 0)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		lg.Warn("redis unavailable, proceeding without cache/blacklist/rate-limit", "error", err)
	}
	permCache := cache.NewPermissionCache(rdb)
	blacklist := cache.NewBlacklist(rdb)

	// --- OpenTelemetry tracing → Jaeger ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdown, tErr := tracing.Init(ctx, "rbac-platform", cfg.JaegerEndpoint)
	if tErr != nil {
		lg.Warn("tracing init failed, continuing without traces", "error", tErr)
	} else {
		defer func() { _ = shutdown(ctx) }()
	}

	// --- Repositories ---
	userRepo := pgrepo.NewUserRepository(db)
	roleRepo := pgrepo.NewRoleRepository(db)
	permRepo := pgrepo.NewPermissionRepository(db)
	tokenRepo := pgrepo.NewRefreshTokenRepository(db)
	outboxRepo := pgrepo.NewOutboxRepository(db)
	auditRepo := pgrepo.NewAuditRepository(db)
	txRunner := pgrepo.NewTxRunner(db)

	// --- Services ---
	jwtMgr := jwtpkg.NewManager(cfg.JWTSecret, cfg.AccessTokenTTL)
	authSvc := auth.NewService(userRepo, roleRepo, tokenRepo, jwtMgr, cfg.RefreshTokenTTL, blacklist)
	rbacSvc := rbac.NewService(userRepo, roleRepo, permRepo, outboxRepo, txRunner, permCache)
	userSvc := user.NewService(userRepo)

	// --- Handlers ---
	authHandler := auth.NewHandler(authSvc)
	rbacHandler := rbac.NewHandler(rbacSvc, rbacSvc.Authorize)
	userHandler := user.NewHandler(userSvc, rbacSvc.Authorize)
	auditHandler := audit.NewHandler(auditRepo, rbacSvc.Authorize)

	// --- RabbitMQ (optional -- event-driven side effects degrade to nothing) ---
	var rmqConn *rabbitmq.Conn
	rmqConn, err = rabbitmq.Connect(cfg.RabbitMQURL)
	if err != nil {
		lg.Warn("rabbitmq unavailable, outbox relay and consumers will not start", "error", err)
	} else {
		defer rmqConn.Close()

		// Start outbox relay
		relay := outbox.NewRelay(outboxRepo, rmqConn.Channel(), lg)
		go relay.Run(ctx)

		// Start consumers
		auditConsumer := audit.NewConsumer(auditRepo)
		notifConsumer := notification.NewConsumer(lg)
		analyticsConsumer := analytics.NewConsumer(lg)

		startConsumer(ctx, rmqConn, "audit.queue", "audit-consumer", lg, auditConsumer.Handle)
		startConsumer(ctx, rmqConn, "notification.queue", "notification-consumer", lg, notifConsumer.Handle)
		startConsumer(ctx, rmqConn, "analytics.queue", "analytics-consumer", lg, analyticsConsumer.Handle)
	}

	// --- HTTP router ---
	router := gin.Default()
	router.Use(middleware.HTTPMetrics())
	router.Use(middleware.Tracing())
	router.Use(middleware.RateLimit(rdb, 100, time.Minute))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/api/v1")

	// Public auth routes (no JWT needed)
	authHandler.RegisterRoutes(v1.Group("/auth"))

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(jwtMgr, blacklist))
	{
		// Logout needs the JWT context for blacklisting, so it lives behind JWTAuth
		authHandler.RegisterProtectedRoutes(protected.Group("/auth"))
		rbacHandler.RegisterRoutes(protected)
		userHandler.RegisterRoutes(protected.Group("/users"))
		auditHandler.RegisterRoutes(protected)
	}

	// --- Graceful shutdown ---
	go func() {
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	lg.Info("shutting down")
	cancel()
}

func startConsumer(ctx context.Context, conn *rabbitmq.Conn, queue, name string, lg *slog.Logger, handler rabbitmq.HandlerFunc) {
	if err := rabbitmq.Consume(ctx, conn.Channel(), queue, name, lg, handler); err != nil {
		lg.Error("failed to start consumer", "queue", queue, "error", err)
	}
}

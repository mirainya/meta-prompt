package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"meta-prompt/internal/config"
	"meta-prompt/internal/handler"
	"meta-prompt/internal/llm"
	"meta-prompt/internal/middleware"
	"meta-prompt/internal/model"
	"meta-prompt/internal/service"
	"meta-prompt/internal/store"
)

//go:embed dist/*
var webFS embed.FS

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	db.AutoMigrate(
		&model.Template{}, &model.History{}, &model.APIKey{},
		&model.User{}, &model.TemplateVersion{},
		&model.ChannelSource{}, &model.ChannelModel{},
	)

	if err := store.SeedDefaults(db); err != nil {
		log.Fatalf("failed to seed defaults: %v", err)
	}

	// stores
	templateStore := store.NewTemplateStore(db)
	historyStore := store.NewHistoryStore(db)
	userStore := store.NewUserStore(db)
	channelStore := store.NewChannelStore(db)
	apiKeyStore := store.NewAPIKeyStore(db)
	templateVersionStore := store.NewTemplateVersionStore(db)

	// 确保 admin 用户的 role
	if adminUser, err := userStore.GetByUsername("admin"); err == nil && adminUser.Role != "admin" {
		userStore.SetRole(adminUser.ID, "admin")
	}

	// ProviderManager
	providerMgr := llm.NewProviderManager(channelStore)

	// default model
	defaultModel := cfg.Defaults.DefaultModel
	if defaultModel == "" {
		defaultModel = cfg.Defaults.LLMProvider // backward compat
	}

	// services
	eventBus := service.NewEventBus()
	webhookSvc := service.NewWebhookService(historyStore)
	analyzer := service.NewAnalyzer(providerMgr)
	architect := service.NewArchitect(providerMgr)
	writer := service.NewWriter(providerMgr)
	reviewer := service.NewReviewer(providerMgr)
	pipeline := service.NewPipeline(analyzer, architect, writer, reviewer, templateStore, historyStore, eventBus, webhookSvc)

	jwtSecret := cfg.Auth.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "meta-prompt-default-secret-change-me"
	}
	authService := service.NewAuthService(userStore, jwtSecret)

	// handlers
	templateHandler := handler.NewTemplateHandler(templateStore, templateVersionStore)
	historyHandler := handler.NewHistoryHandler(historyStore)
	generateHandler := handler.NewGenerateHandler(pipeline, defaultModel, userStore, historyStore, channelStore)
	authHandler := handler.NewAuthHandler(authService, userStore)
	adminHandler := handler.NewAdminHandler(userStore, historyStore)
	channelHandler := handler.NewChannelHandler(channelStore, userStore)
	sseHandler := handler.NewSSEHandler(eventBus, historyStore)
	exportHandler := handler.NewExportHandler(historyStore)
	openHandler := handler.NewOpenHandler(pipeline, historyStore, userStore, channelStore, defaultModel)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeyStore, historyStore)
	docsHandler := handler.NewDocsHandler()

	r := gin.Default()

	// 结构化日志
	r.Use(middleware.Logger())

	// Rate Limiting
	limiter := middleware.NewRateLimiter(rate.Limit(10), 20)
	r.Use(limiter.Middleware())

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	api := r.Group("/api/v1")

	// 公开路由
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// 需要认证的路由
	protected := api.Group("")
	protected.Use(middleware.JWTAuth(authService))
	{
		protected.GET("/user/me", authHandler.Me)
		protected.GET("/models", channelHandler.ListAvailableModels)

		protected.POST("/generate", generateHandler.Generate)
		protected.POST("/histories/:id/cancel", generateHandler.Cancel)
		protected.GET("/histories/:id/stream", sseHandler.Stream)
		protected.GET("/histories/:id/export", exportHandler.Export)

		protected.POST("/templates", templateHandler.Create)
		protected.GET("/templates", templateHandler.List)
		protected.GET("/templates/:id", templateHandler.Get)
		protected.PUT("/templates/:id", templateHandler.Update)
		protected.DELETE("/templates/:id", templateHandler.Delete)

		protected.GET("/histories", historyHandler.List)
		protected.GET("/histories/running", historyHandler.Running)
		protected.GET("/histories/:id", historyHandler.Get)

		protected.GET("/api-keys", apiKeyHandler.List)
		protected.POST("/api-keys", apiKeyHandler.Create)
		protected.DELETE("/api-keys/:id", apiKeyHandler.Revoke)
		protected.GET("/api-keys/:id/stats", apiKeyHandler.Stats)
	}

	// Admin 路由
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(authService), middleware.AdminAuth(userStore))
	{
		admin.GET("/dashboard", adminHandler.Dashboard)

		// 渠道管理
		admin.GET("/channels/sources", channelHandler.ListSources)
		admin.POST("/channels/sources", channelHandler.CreateSource)
		admin.PUT("/channels/sources/:id", channelHandler.UpdateSource)
		admin.DELETE("/channels/sources/:id", channelHandler.DeleteSource)
		admin.POST("/channels/sources/:id/sync", channelHandler.SyncSource)
		admin.GET("/channels/models", channelHandler.ListModels)
		admin.PUT("/channels/models/:id", channelHandler.UpdateModel)

		admin.GET("/users", adminHandler.ListUsers)
		admin.PUT("/users/:id/credits", adminHandler.SetUserCredits)
		admin.PUT("/users/:id/role", adminHandler.SetUserRole)
		admin.PUT("/users/:id/reset-password", adminHandler.ResetUserPassword)
		admin.PUT("/users/:id/models", adminHandler.SetUserModels)
		admin.GET("/templates", templateHandler.List)
		admin.GET("/templates/:id", templateHandler.Get)
		admin.PUT("/templates/:id", templateHandler.Update)
		admin.GET("/templates/:id/versions", templateHandler.ListVersions)
		admin.POST("/templates/:id/rollback", templateHandler.Rollback)
		admin.GET("/api-keys", apiKeyHandler.AdminList)
		admin.POST("/api-keys", apiKeyHandler.AdminCreate)
		admin.DELETE("/api-keys/:id", apiKeyHandler.AdminRevoke)
	}

	// Open API 文档（无需鉴权）
	r.GET("/open/v1/docs", docsHandler.OpenAPISpec)

	// Open API 路由（外部平台接入）
	open := r.Group("/open/v1")
	open.Use(middleware.APIKeyAuth(apiKeyStore))
	{
		open.POST("/generate", openHandler.Generate)
		open.GET("/tasks/:id", openHandler.GetTask)
		open.GET("/tasks/:id/stream", sseHandler.Stream)
		open.GET("/tasks/:id/export", exportHandler.Export)
		open.POST("/tasks/:id/cancel", openHandler.CancelTask)
		open.GET("/models", channelHandler.ListAvailableModels)
	}

	// 静态文件服务 — SPA fallback
	distFS, _ := fs.Sub(webFS, "dist")
	fileServer := http.FileServer(http.FS(distFS))
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/open/") {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		f, err := fs.Stat(distFS, strings.TrimPrefix(path, "/"))
		if err == nil && !f.IsDir() {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}
		c.FileFromFS("/", http.FS(distFS))
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	slog.Info("server starting", "addr", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

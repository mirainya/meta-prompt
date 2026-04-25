package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	db.AutoMigrate(&model.Template{}, &model.History{}, &model.APIKey{}, &model.User{}, &model.LLMConfig{})

	if err := store.SeedDefaults(db); err != nil {
		log.Fatalf("failed to seed defaults: %v", err)
	}

	// stores
	templateStore := store.NewTemplateStore(db)
	historyStore := store.NewHistoryStore(db)
	userStore := store.NewUserStore(db)
	llmConfigStore := store.NewLLMConfigStore(db)

	// 确保 admin 用户的 role
	if adminUser, err := userStore.GetByUsername("admin"); err == nil && adminUser.Role != "admin" {
		userStore.SetRole(adminUser.ID, "admin")
	}

	// 从 config.yaml 导入 LLM 配置到数据库（仅首次）
	seedLLMConfigs(cfg, llmConfigStore)

	// ProviderManager
	providerMgr := llm.NewProviderManager()
	dbConfigs, _ := llmConfigStore.List()
	handler.RebuildAllProviders(providerMgr, dbConfigs)

	// services
	analyzer := service.NewAnalyzer(providerMgr)
	architect := service.NewArchitect(providerMgr)
	writer := service.NewWriter(providerMgr)
	reviewer := service.NewReviewer(providerMgr)
	pipeline := service.NewPipeline(analyzer, architect, writer, reviewer, templateStore, historyStore)

	jwtSecret := cfg.Auth.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "meta-prompt-default-secret-change-me"
	}
	authService := service.NewAuthService(userStore, jwtSecret)

	// handlers
	templateHandler := handler.NewTemplateHandler(templateStore)
	historyHandler := handler.NewHistoryHandler(historyStore)
	generateHandler := handler.NewGenerateHandler(pipeline, cfg.Defaults.LLMProvider, userStore, historyStore)
	authHandler := handler.NewAuthHandler(authService, userStore)
	adminHandler := handler.NewAdminHandler(userStore, llmConfigStore, historyStore, providerMgr)
	providerHandler := handler.NewProviderHandler(llmConfigStore)

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
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
		protected.GET("/providers", providerHandler.ListEnabled)

		protected.POST("/generate", generateHandler.Generate)
		protected.POST("/histories/:id/cancel", generateHandler.Cancel)

		protected.POST("/templates", templateHandler.Create)
		protected.GET("/templates", templateHandler.List)
		protected.GET("/templates/:id", templateHandler.Get)
		protected.PUT("/templates/:id", templateHandler.Update)
		protected.DELETE("/templates/:id", templateHandler.Delete)

		protected.GET("/histories", historyHandler.List)
		protected.GET("/histories/running", historyHandler.Running)
		protected.GET("/histories/:id", historyHandler.Get)
	}

	// Admin 路由
	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(authService), middleware.AdminAuth(userStore))
	{
		admin.GET("/dashboard", adminHandler.Dashboard)
		admin.GET("/llm-configs", adminHandler.ListLLMConfigs)
		admin.POST("/llm-configs", adminHandler.CreateLLMConfig)
		admin.PUT("/llm-configs/:provider", adminHandler.UpdateLLMConfig)
		admin.DELETE("/llm-configs/:provider", adminHandler.DeleteLLMConfig)
		admin.POST("/llm-configs/:provider/test", adminHandler.TestLLMConfig)
		admin.GET("/users", adminHandler.ListUsers)
		admin.PUT("/users/:id/credits", adminHandler.SetCredits)
		admin.PUT("/users/:id/toggle", adminHandler.ToggleUser)
		admin.PUT("/users/:id/reset-password", adminHandler.ResetPassword)
		admin.GET("/templates", templateHandler.List)
		admin.GET("/templates/:id", templateHandler.Get)
		admin.PUT("/templates/:id", templateHandler.Update)
	}

	// 静态文件服务 — SPA fallback
	distFS, _ := fs.Sub(webFS, "dist")
	fileServer := http.FileServer(http.FS(distFS))
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") {
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
	log.Printf("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

// seedLLMConfigs 从 config.yaml 导入 LLM 配置到数据库（仅当数据库中无记录时）
func seedLLMConfigs(cfg *config.Config, llmConfigStore *store.LLMConfigStore) {
	existing, _ := llmConfigStore.List()
	if len(existing) > 0 {
		return
	}

	defaults := map[string]struct {
		baseURL string
		typ     string
	}{
		"claude": {"https://api.anthropic.com", "claude"},
		"openai": {"https://api.openai.com", "openai_compatible"},
		"gemini": {"https://generativelanguage.googleapis.com", "gemini"},
	}

	for name, llmCfg := range cfg.LLM {
		d := defaults[name]
		llmConfigStore.Upsert(&model.LLMConfig{
			Provider:  name,
			Type:      d.typ,
			APIKey:    llmCfg.APIKey,
			BaseURL:   d.baseURL,
			Model:     llmCfg.Model,
			MaxTokens: llmCfg.MaxTokens,
			Enabled:   llmCfg.APIKey != "",
		})
	}
}

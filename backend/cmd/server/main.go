package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/kunal/life-log/backend/configs"
	"github.com/kunal/life-log/backend/internal/analytics"
	"github.com/kunal/life-log/backend/internal/auth"
	"github.com/kunal/life-log/backend/internal/database"
	"github.com/kunal/life-log/backend/internal/entries"
	"github.com/kunal/life-log/backend/internal/goals"
	"github.com/kunal/life-log/backend/internal/habits"
	"github.com/kunal/life-log/backend/internal/middleware"
	"github.com/kunal/life-log/backend/internal/mqtt"
	"github.com/kunal/life-log/backend/internal/notifications"
	"github.com/kunal/life-log/backend/internal/users"
)

func main() {
	cfg := configs.Load()

	gin.SetMode(cfg.Server.Mode)

	db := database.Init(&cfg.Database)
	database.Migrate()

	authSvc := auth.NewService(db, &cfg.Auth)
	userSvc := users.NewService(db)
	entrySvc := entries.NewService(db)
	habitSvc := habits.NewService(db)
	goalSvc := goals.NewService(db)
	analyticsSvc := analytics.NewService(db)
	notifSvc := notifications.NewService(db, &cfg.Ntfy)
	mqttSvc := mqtt.NewService(db, entrySvc, &cfg.MQTT)

	rateLimiter := middleware.NewRateLimiter(&cfg.RateLimit)

	r := gin.New()
	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/healthz", "/readyz"},
	}))
	r.Use(gin.Recovery())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     cfg.CORS.AllowedMethods,
		AllowHeaders:     cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	}))

	r.Use(rateLimiter.Middleware())
	r.Use(middleware.SecureHeaders())
	r.Use(middleware.RequestID())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC()})
	})
	r.GET("/readyz", func(c *gin.Context) {
		dbInst, _ := db.DB()
		if err := dbInst.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	api := r.Group("/api/v1")

	mw := middleware.New(authSvc)

	auth.RegisterRoutes(api, authSvc, mw.RequireAuth())
	users.RegisterRoutes(api, userSvc, mw.RequireAuth())
	entries.RegisterRoutes(api, entrySvc, mw.RequireAuth())
	habits.RegisterRoutes(api, habitSvc, mw.RequireAuth())
	goals.RegisterRoutes(api, goalSvc, mw.RequireAuth())
	analytics.RegisterRoutes(api, analyticsSvc, mw.RequireAuth())
	notifications.RegisterRoutes(api, notifSvc, mw.RequireAuth())
	mqtt.RegisterRoutes(api, mqttSvc, mw.RequireAdmin())

	public := api.Group("/public")
	{
		public.GET("/feed", entrySvc.HandlePublicFeed)
		public.GET("/timeline", entrySvc.HandlePublicTimeline)
		public.GET("/stats", entrySvc.HandlePublicStats)
		public.GET("/heatmap", entrySvc.HandlePublicHeatmap)
	}

	r.GET("/rss.xml", entrySvc.HandleRSS)
	r.GET("/feed.json", entrySvc.HandleJSONFeed)

	admin := api.Group("/admin")
	admin.Use(mw.RequireAdmin())
	{
		admin.GET("/users", userSvc.HandleAdminListUsers)
		admin.GET("/audit-logs", authSvc.HandleAuditLogs)
		admin.GET("/stats", analyticsSvc.HandleSystemStats)
	}

	if cfg.MQTT.Broker != "" {
		go mqttSvc.Start()
	}

	if cfg.Ntfy.Enabled {
		notifSvc.StartDailyReminder()
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	log.Printf("LifeLog server starting on :%s", cfg.Server.Port)
	log.Printf("Database driver: %s", cfg.Database.Driver)
	log.Printf("Base URL: %s", cfg.Server.BaseURL)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

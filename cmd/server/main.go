package main

import (
	"Dimidroll06/url-link-shortener/internal/adapters/cache"
	"Dimidroll06/url-link-shortener/internal/adapters/handlers"
	"Dimidroll06/url-link-shortener/internal/adapters/repository"
	"Dimidroll06/url-link-shortener/internal/adapters/server"
	"Dimidroll06/url-link-shortener/internal/config"
	"Dimidroll06/url-link-shortener/internal/core/services"
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	rootContext, cancel := context.WithCancel(context.Background())

	cfg := config.Load()

	logger, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("failed to init logger %v", err.Error())
	}

	db, err := pgxpool.New(rootContext, cfg.DatabaseURL())
	if err != nil {
		logger.Fatal("failed to init db", zap.Error(err))
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURL(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := rdb.Ping(rootContext).Err(); err != nil {
		logger.Fatal("failed to init redis", zap.Error(err))
	}

	router := setupRouter(db, rdb, logger, cfg)

	srv := server.NewServer(
		router,
		"8080",
		db,
		rdb,
		logger,
		30*time.Second,
	)

	if err := srv.Run(rootContext); err != nil {
		cancel()
		if err == context.Canceled {
			logger.Info("server stopped gracefully")
			return
		}
		logger.Fatal("server failed to start", zap.Error(err))
		os.Exit(1)
	}
}

func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var config zap.Config

	if cfg.IsDevelopment() {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	levels := map[string]int{
		"debug": -1,
		"info":  0,
		"warn":  1,
		"error": 2,
		"panic": 3,
		"fatal": 5,
	}

	config.Level = zap.NewAtomicLevelAt(zapcore.Level(levels[cfg.LogLevel]))
	config.Encoding = cfg.LogFormat

	return config.Build()
}

func setupRouter(db *pgxpool.Pool, rdb *redis.Client, logger *zap.Logger, cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.GinMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	urlRepo := repository.NewURLRepository(db)
	statsRepo := repository.NewStatsRepository(db)

	urlCache := cache.NewURLCache(rdb, "urlshortener")
	statsCache := cache.NewStatsCache(rdb, "urlshortener")

	urlService := services.NewURLService(urlRepo, urlCache, statsCache, logger, cfg.URLExpirationDays)
	statsService := services.NewStatsService(statsRepo, statsCache, urlRepo, logger)

	urlHandler := handlers.NewURLHandler(urlService, statsService, logger, cfg.BaseURL)

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy", "error": "database"})
			return
		}

		if err := rdb.Ping(ctx).Err(); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy", "error": "redis"})
			return
		}

		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now()})
	})

	urlHandler.RegisterRoutes(r)

	return r
}

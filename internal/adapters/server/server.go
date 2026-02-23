package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Server struct {
	httpServer   *http.Server
	db           *pgxpool.Pool
	redis        *redis.Client
	logger       *zap.Logger
	shutdownTime time.Duration
}

func NewServer(
	router *gin.Engine,
	port string,
	db *pgxpool.Pool,
	redis *redis.Client,
	logger *zap.Logger,
	shutdownTime time.Duration,
) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         ":" + port,
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  30 * time.Second,
		},
		db:           db,
		redis:        redis,
		logger:       logger,
		shutdownTime: shutdownTime,
	}
}

func (s *Server) Run(ctx context.Context) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s.logger.Info("server starting", zap.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("server failed to start", zap.Error(err))
		}
	}()

	<-quit
	s.logger.Info("server shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTime)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("server failed to shutdown gracefully", zap.Error(err))
	}
	s.logger.Info("server stopped gracefully")

	if s.db != nil {
		if err := s.closeDB(shutdownCtx); err != nil {
			s.logger.Error("failed to close db", zap.Error(err))
		}
	}

	if s.redis != nil {
		if err := s.closeRedis(shutdownCtx); err != nil {
			s.logger.Error("failed to close redis", zap.Error(err))
		}
	}

	return nil
}

func (s *Server) closeDB(ctx context.Context) error {
	done := make(chan error, 1)

	go func() {
		s.db.Close()
		done <- nil
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("db close timeout %w", ctx.Err())
	case err := <-done:
		return err
	}
}

func (s *Server) closeRedis(ctx context.Context) error {
	done := make(chan error, 1)

	go func() {
		done <- s.redis.Close()
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("redis close timeout %w", ctx.Err())
	case err := <-done:
		return err
	}
}

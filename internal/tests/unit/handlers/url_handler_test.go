package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"Dimidroll06/url-link-shortener/internal/adapters/handlers"
	"Dimidroll06/url-link-shortener/internal/core/domain"
	servererrors "Dimidroll06/url-link-shortener/internal/core/errors"
	mockServices "Dimidroll06/url-link-shortener/internal/tests/mock/services"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestURLHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		requestBody      string
		setupMocks       func(*mockServices.MockURLService, *mockServices.MockStatsService)
		wantStatus       int
		wantBodyContains string
	}{
		{
			name:        "success_create",
			requestBody: `{"url": "https://example.com"}`,
			setupMocks: func(urlSvc *mockServices.MockURLService, statsSvc *mockServices.MockStatsService) {
				urlSvc.CreateFunc = func(ctx context.Context, originalURL string) (*domain.URL, error) {
					assert.Equal(t, "https://example.com", originalURL)
					return &domain.URL{
						ID:          "test-id",
						OriginalURL: "https://example.com",
						ShortCode:   "abc123",
						CreatedAt:   time.Now().UTC(),
						IsActive:    true,
					}, nil
				}
			},
			wantStatus:       http.StatusCreated,
			wantBodyContains: `"short_code":"abc123"`,
		},
		{
			name:             "error_invalid_json",
			requestBody:      `{"url": invalid}`,
			setupMocks:       func(urlSvc *mockServices.MockURLService, statsSvc *mockServices.MockStatsService) {},
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: `"error":"invalid request body"`,
		},
		{
			name:             "error_empty_url",
			requestBody:      `{"url": ""}`,
			setupMocks:       func(urlSvc *mockServices.MockURLService, statsSvc *mockServices.MockStatsService) {},
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: `"error"`,
		},
		{
			name:        "error_invalid_scheme",
			requestBody: `{"url": "ftp://example.com"}`,
			setupMocks: func(urlSvc *mockServices.MockURLService, statsSvc *mockServices.MockStatsService) {
				urlSvc.CreateFunc = func(ctx context.Context, originalURL string) (*domain.URL, error) {
					return nil, servererrors.ErrInvalidURLScheme
				}
			},
			wantStatus:       http.StatusBadRequest,
			wantBodyContains: `"error":"invalid URL scheme`,
		},
		{
			name:        "error_short_code_exists",
			requestBody: `{"url": "https://example.com"}`,
			setupMocks: func(urlSvc *mockServices.MockURLService, statsSvc *mockServices.MockStatsService) {
				urlSvc.CreateFunc = func(ctx context.Context, originalURL string) (*domain.URL, error) {
					return nil, servererrors.ErrShortCodeExists
				}
			},
			wantStatus:       http.StatusConflict,
			wantBodyContains: `"error":"short code already exists"`,
		},
		{
			name:        "error_internal",
			requestBody: `{"url": "https://example.com"}`,
			setupMocks: func(urlSvc *mockServices.MockURLService, statsSvc *mockServices.MockStatsService) {
				urlSvc.CreateFunc = func(ctx context.Context, originalURL string) (*domain.URL, error) {
					return nil, errors.New("db connection failed")
				}
			},
			wantStatus:       http.StatusInternalServerError,
			wantBodyContains: `"error":"internal server error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockURLService := mockServices.NewMockURLService()
			mockStatsService := mockServices.NewMockStatsService()
			logger := zap.NewNop()

			tt.setupMocks(mockURLService, mockStatsService)

			handler := handlers.NewURLHandler(mockURLService, mockStatsService, logger, "http://localhost:8080")

			r := gin.New()
			r.Use(gin.Recovery())
			handler.RegisterRoutes(r)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", bytes.NewReader([]byte(tt.requestBody)))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBodyContains != "" {
				assert.Contains(t, w.Body.String(), tt.wantBodyContains)
			}
		})
	}
}

func TestURLHandler_Redirect(t *testing.T) {
	t.Parallel()

	t.Run("success_redirect", func(t *testing.T) {
		t.Parallel()

		mockURLService := mockServices.NewMockURLService()
		mockStatsService := mockServices.NewMockStatsService()
		logger := zap.NewNop()

		mockURLService.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			assert.Equal(t, "abc123", code)
			return &domain.URL{
				ID:          "test-id",
				OriginalURL: "https://example.com/target",
				ShortCode:   "abc123",
				IsActive:    true,
			}, nil
		}

		handler := handlers.NewURLHandler(mockURLService, mockStatsService, logger, "http://localhost:8080")

		r := gin.New()
		r.Use(gin.Recovery())
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "https://example.com/target", w.Header().Get("Location"))
	})

	t.Run("error_url_not_found", func(t *testing.T) {
		t.Parallel()

		mockURLService := mockServices.NewMockURLService()
		mockStatsService := mockServices.NewMockStatsService()
		logger := zap.NewNop()

		mockURLService.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLNotFound
		}

		handler := handlers.NewURLHandler(mockURLService, mockStatsService, logger, "http://localhost:8080")

		r := gin.New()
		r.Use(gin.Recovery())
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"link not found"`)
	})

	t.Run("error_url_expired", func(t *testing.T) {
		t.Parallel()

		mockURLService := mockServices.NewMockURLService()
		mockStatsService := mockServices.NewMockStatsService()
		logger := zap.NewNop()

		mockURLService.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLExpired
		}

		handler := handlers.NewURLHandler(mockURLService, mockStatsService, logger, "http://localhost:8080")

		r := gin.New()
		r.Use(gin.Recovery())
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/expired", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"link not found"`)
	})

	t.Run("error_url_inactive", func(t *testing.T) {
		t.Parallel()

		mockURLService := mockServices.NewMockURLService()
		mockStatsService := mockServices.NewMockStatsService()
		logger := zap.NewNop()

		mockURLService.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLInactive
		}

		handler := handlers.NewURLHandler(mockURLService, mockStatsService, logger, "http://localhost:8080")

		r := gin.New()
		r.Use(gin.Recovery())
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/inactive", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"link not found"`)
	})

	t.Run("error_internal", func(t *testing.T) {
		t.Parallel()

		mockURLService := mockServices.NewMockURLService()
		mockStatsService := mockServices.NewMockStatsService()
		logger := zap.NewNop()

		mockURLService.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, errors.New("db error")
		}

		handler := handlers.NewURLHandler(mockURLService, mockStatsService, logger, "http://localhost:8080")

		r := gin.New()
		r.Use(gin.Recovery())
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"internal server error"`)
	})
}

func TestURLHandler_GetStats(t *testing.T) {
	t.Parallel()

	t.Run("success_get_stats", func(t *testing.T) {
		t.Parallel()

		mockURLService := mockServices.NewMockURLService()
		mockStatsService := mockServices.NewMockStatsService()
		logger := zap.NewNop()

		expectedURL := &domain.URL{
			ID:          "test-id",
			OriginalURL: "https://example.com",
			ShortCode:   "abc123",
			CreatedAt:   time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			ExpiresAt:   func() *time.Time { t := time.Date(2024, 2, 15, 10, 30, 0, 0, time.UTC); return &t }(),
			IsActive:    true,
		}

		mockURLService.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return expectedURL, nil
		}

		mockURLService.GetStatsFunc = func(ctx context.Context, code string) (int64, error) {
			return 42, nil
		}

		handler := handlers.NewURLHandler(mockURLService, mockStatsService, logger, "http://localhost:8080")

		r := gin.New()
		r.Use(gin.Recovery())
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/abc123/stats", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "abc123", resp["short_code"])
		assert.Equal(t, "https://example.com", resp["original_url"])
		assert.EqualValues(t, 42, resp["total_accesses"])
	})

	t.Run("error_url_not_found", func(t *testing.T) {
		t.Parallel()

		mockURLService := mockServices.NewMockURLService()
		mockStatsService := mockServices.NewMockStatsService()
		logger := zap.NewNop()

		mockURLService.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return nil, servererrors.ErrURLNotFound
		}

		handler := handlers.NewURLHandler(mockURLService, mockStatsService, logger, "http://localhost:8080")

		r := gin.New()
		r.Use(gin.Recovery())
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent/stats", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"link not found"`)
	})

	t.Run("error_stats_unavailable", func(t *testing.T) {
		t.Parallel()

		mockURLService := mockServices.NewMockURLService()
		mockStatsService := mockServices.NewMockStatsService()
		logger := zap.NewNop()

		mockURLService.GetByShortCodeFunc = func(ctx context.Context, code string) (*domain.URL, error) {
			return &domain.URL{ShortCode: code, IsActive: true}, nil
		}

		mockURLService.GetStatsFunc = func(ctx context.Context, code string) (int64, error) {
			return 0, servererrors.ErrStatsUnavailable
		}

		handler := handlers.NewURLHandler(mockURLService, mockStatsService, logger, "http://localhost:8080")

		r := gin.New()
		r.Use(gin.Recovery())
		handler.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/abc123/stats", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), `"error":"failed to get statistics"`)
	})
}

func TestURLHandler_RegisterRoutes(t *testing.T) {
	t.Parallel()

	mockURLService := mockServices.NewMockURLService()
	mockStatsService := mockServices.NewMockStatsService()
	logger := zap.NewNop()

	handler := handlers.NewURLHandler(mockURLService, mockStatsService, logger, "http://localhost:8080")

	r := gin.New()
	handler.RegisterRoutes(r)

	routes := r.Routes()

	var foundCreate, foundStats, foundRedirect bool
	for _, route := range routes {
		if route.Method == "POST" && route.Path == "/api/v1/shorten" {
			foundCreate = true
		}
		if route.Method == "GET" && route.Path == "/api/v1/:code/stats" {
			foundStats = true
		}
		if route.Method == "GET" && route.Path == "/:code" {
			foundRedirect = true
		}
	}

	assert.True(t, foundCreate, "POST /api/v1/shorten not registered")
	assert.True(t, foundStats, "GET /api/v1/:code/stats not registered")
	assert.True(t, foundRedirect, "GET /:code not registered")
}

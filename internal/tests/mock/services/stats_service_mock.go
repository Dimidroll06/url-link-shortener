package services

import (
	"context"
)

type MockStatsService struct {
	GetDetailedStatsFunc func(ctx context.Context, code string) (map[string]interface{}, error)
	RecordAccessFunc     func(ctx context.Context, code, ipAddress, userAgent, referer string) error
}

func NewMockStatsService() *MockStatsService {
	return &MockStatsService{
		GetDetailedStatsFunc: func(ctx context.Context, code string) (map[string]interface{}, error) {
			return nil, nil
		},
		RecordAccessFunc: func(ctx context.Context, code, ipAddress, userAgent, referer string) error {
			return nil
		},
	}
}

func (m *MockStatsService) GetDetailedStats(ctx context.Context, code string) (map[string]interface{}, error) {
	return m.GetDetailedStatsFunc(ctx, code)
}

func (m *MockStatsService) RecordAccess(ctx context.Context, code, ipAddress, userAgent, referer string) error {
	return m.RecordAccessFunc(ctx, code, ipAddress, userAgent, referer)
}

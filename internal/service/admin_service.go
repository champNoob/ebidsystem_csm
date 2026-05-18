package service

import (
	"context"
	"ebidsystem_csm/internal/repository"
	"ebidsystem_csm/internal/repository/dto"
)

type AdminService struct {
	repo repository.AdminRepository
}

func NewAdminService(repo repository.AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) GetDashboard(ctx context.Context) (map[string]interface{}, error) {
	totalOrders, totalTrades, totalVolume, totalTurnover, err :=
		s.repo.GetGlobalStats(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_orders":   totalOrders,
		"total_trades":   totalTrades,
		"total_volume":   totalVolume,
		"total_turnover": totalTurnover,
	}, nil
}

func (s *AdminService) GetSymbolStats(ctx context.Context) ([]dto.SymbolStat, error) {
	return s.repo.GetSymbolStats(ctx)
}

func (s *AdminService) GetOrderStatusStats(ctx context.Context) ([]dto.OrderStatusStat, error) {
	return s.repo.GetOrderStatusStats(ctx)
}

func (s *AdminService) GetUserRoleStats(ctx context.Context) ([]dto.UserRoleStat, error) {
	return s.repo.GetUserRoleStats(ctx)
}

func (s *AdminService) GetUserRanking(ctx context.Context) ([]dto.UserRank, error) {
	return s.repo.GetUserRanking(ctx, 10)
}

func (s *AdminService) GetRecentTrades(ctx context.Context, limit int) ([]dto.RecentTrade, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.GetRecentTrades(ctx, limit)
}

func (s *AdminService) GetTradeTimeline(ctx context.Context) ([]dto.TradeTimelinePoint, error) {
	return s.repo.GetTradeTimeline(ctx)
}

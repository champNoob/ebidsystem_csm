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

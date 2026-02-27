package services

import (
	"context"
	"fmt"

	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/repository"
)

// DashboardStats aggregates all data needed for the admin dashboard.
type DashboardStats struct {
	TotalUsers     int64
	TotalTours     int64
	TodayBookings  int64
	MonthRevenue   float64
	RecentBookings []models.Booking
	PendingReviews []models.Review
}

type StatsService struct {
	repo *repository.StatsRepository
}

func NewStatsService(repo *repository.StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

// GetDashboardStats collects all stats required by the admin dashboard.
// On query failure the error is wrapped and returned; the handler falls back to zero values.
func (s *StatsService) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}
	var err error

	if stats.TotalUsers, err = s.repo.CountUsers(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", messages.ErrCtxCountUsers, err)
	}

	if stats.TotalTours, err = s.repo.CountActiveTours(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", messages.ErrCtxCountTours, err)
	}

	if stats.TodayBookings, err = s.repo.CountTodayBookings(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", messages.ErrCtxCountTodayBookings, err)
	}

	if stats.MonthRevenue, err = s.repo.SumMonthRevenue(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", messages.ErrCtxSumMonthRevenue, err)
	}

	if stats.RecentBookings, err = s.repo.RecentBookings(ctx, 5); err != nil {
		return nil, fmt.Errorf("%s: %w", messages.ErrCtxRecentBookings, err)
	}

	if stats.PendingReviews, err = s.repo.PendingReviews(ctx, 5); err != nil {
		return nil, fmt.Errorf("%s: %w", messages.ErrCtxPendingReviews, err)
	}

	return stats, nil
}

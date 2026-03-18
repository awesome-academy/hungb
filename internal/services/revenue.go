package services

import (
	"context"
	"fmt"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/repository"
)

type RevenueStats struct {
	TotalRevenue     float64
	TotalBookings    int64
	SuccessPayments  int64
	AvgBookingValue  float64
	MonthlyBreakdown []repository.MonthlyRevenue
	TopTours         []repository.TourRevenue
	BookingsByStatus []repository.BookingStatusCount
}

type RevenueService struct {
	repo repository.RevenueRepo
}

func NewRevenueService(repo repository.RevenueRepo) *RevenueService {
	return &RevenueService{repo: repo}
}

func (s *RevenueService) GetRevenueStats(ctx context.Context, filter repository.RevenueFilter) (*RevenueStats, error) {
	stats := &RevenueStats{}
	var err error

	if stats.TotalRevenue, err = s.repo.TotalRevenue(ctx, filter); err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueServiceGet, err)
	}

	if stats.TotalBookings, err = s.repo.TotalBookings(ctx, filter); err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueServiceGet, err)
	}

	if stats.SuccessPayments, err = s.repo.SuccessPaymentCount(ctx, filter); err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueServiceGet, err)
	}

	if stats.TotalBookings > 0 {
		stats.AvgBookingValue = stats.TotalRevenue / float64(stats.TotalBookings)
	}

	if stats.MonthlyBreakdown, err = s.repo.MonthlyBreakdown(ctx, filter); err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueServiceGet, err)
	}

	if stats.TopTours, err = s.repo.TopToursByRevenue(ctx, filter, 10); err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueServiceGet, err)
	}

	if stats.BookingsByStatus, err = s.repo.BookingsByStatus(ctx, filter); err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueServiceGet, err)
	}

	return stats, nil
}

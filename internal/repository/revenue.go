package repository

import (
	"context"
	"fmt"
	"time"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/models"

	"gorm.io/gorm"
)

type MonthlyRevenue struct {
	Month        string
	PaymentCount int64
	Total        float64
}

type TourRevenue struct {
	TourID       uint
	TourTitle    string
	BookingCount int64
	Total        float64
}

type BookingStatusCount struct {
	Status string
	Count  int64
}

type RevenueFilter struct {
	DateFrom time.Time
	DateTo   time.Time
	TourID   uint
}

type RevenueRepo interface {
	TotalRevenue(ctx context.Context, filter RevenueFilter) (float64, error)
	TotalBookings(ctx context.Context, filter RevenueFilter) (int64, error)
	SuccessPaymentCount(ctx context.Context, filter RevenueFilter) (int64, error)
	MonthlyBreakdown(ctx context.Context, filter RevenueFilter) ([]MonthlyRevenue, error)
	TopToursByRevenue(ctx context.Context, filter RevenueFilter, limit int) ([]TourRevenue, error)
	BookingsByStatus(ctx context.Context, filter RevenueFilter) ([]BookingStatusCount, error)
}

// revenueRepository requires PostgreSQL
type revenueRepository struct {
	db *gorm.DB
}

func NewRevenueRepository(db *gorm.DB) RevenueRepo {
	return &revenueRepository{db: db}
}

func (r *revenueRepository) applyPaymentDateFilter(query *gorm.DB, filter RevenueFilter) *gorm.DB {
	query = query.Where("payments.status = ?", constants.PaymentStatusSuccess)
	if !filter.DateFrom.IsZero() {
		query = query.Where("payments.paid_at >= ?", filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		query = query.Where("payments.paid_at <= ?", filter.DateTo)
	}
	return query
}

func (r *revenueRepository) applyPaymentFilter(query *gorm.DB, filter RevenueFilter) *gorm.DB {
	query = r.applyPaymentDateFilter(query, filter)
	if filter.TourID > 0 {
		query = query.Joins("JOIN bookings ON bookings.id = payments.booking_id").
			Where("bookings.tour_id = ?", filter.TourID)
	}
	return query
}

func (r *revenueRepository) applyBookingFilter(query *gorm.DB, filter RevenueFilter) *gorm.DB {
	if !filter.DateFrom.IsZero() || !filter.DateTo.IsZero() {
		query = query.Joins("JOIN payments ON payments.booking_id = bookings.id AND payments.status = ?", constants.PaymentStatusSuccess)
		if !filter.DateFrom.IsZero() {
			query = query.Where("payments.paid_at >= ?", filter.DateFrom)
		}
		if !filter.DateTo.IsZero() {
			query = query.Where("payments.paid_at <= ?", filter.DateTo)
		}
	}
	if filter.TourID > 0 {
		query = query.Where("bookings.tour_id = ?", filter.TourID)
	}
	return query
}

func (r *revenueRepository) TotalRevenue(ctx context.Context, filter RevenueFilter) (float64, error) {
	var total float64
	query := r.db.WithContext(ctx).Model(&models.Payment{})
	query = r.applyPaymentFilter(query, filter)
	if err := query.Select("COALESCE(SUM(amount), 0)").Scan(&total).Error; err != nil {
		return 0, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueTotalRevenue, err)
	}
	return total, nil
}

func (r *revenueRepository) TotalBookings(ctx context.Context, filter RevenueFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Booking{}).Table("bookings")
	query = r.applyBookingFilter(query, filter)
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueTotalBookings, err)
	}
	return count, nil
}

func (r *revenueRepository) SuccessPaymentCount(ctx context.Context, filter RevenueFilter) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.Payment{})
	query = r.applyPaymentFilter(query, filter)
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueSuccessPayments, err)
	}
	return count, nil
}

func (r *revenueRepository) MonthlyBreakdown(ctx context.Context, filter RevenueFilter) ([]MonthlyRevenue, error) {
	var results []MonthlyRevenue

	query := r.db.WithContext(ctx).Model(&models.Payment{})
	query = r.applyPaymentFilter(query, filter)

	if err := query.
		Select("TO_CHAR(payments.paid_at, 'YYYY-MM') AS month, COUNT(*) AS payment_count, COALESCE(SUM(amount), 0) AS total").
		Group("TO_CHAR(payments.paid_at, 'YYYY-MM')").
		Order("month DESC").
		Limit(12).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueMonthlyBreakdown, err)
	}
	return results, nil
}

func (r *revenueRepository) TopToursByRevenue(ctx context.Context, filter RevenueFilter, limit int) ([]TourRevenue, error) {
	var results []TourRevenue

	query := r.db.WithContext(ctx).Model(&models.Payment{}).
		Joins("JOIN bookings ON bookings.id = payments.booking_id").
		Joins("JOIN tours ON tours.id = bookings.tour_id")
	query = r.applyPaymentDateFilter(query, filter)
	if filter.TourID > 0 {
		query = query.Where("bookings.tour_id = ?", filter.TourID)
	}

	if err := query.
		Select("bookings.tour_id, tours.title AS tour_title, COUNT(DISTINCT bookings.id) AS booking_count, COALESCE(SUM(payments.amount), 0) AS total").
		Group("bookings.tour_id, tours.title").
		Order("total DESC").
		Limit(limit).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueByTour, err)
	}
	return results, nil
}

func (r *revenueRepository) BookingsByStatus(ctx context.Context, filter RevenueFilter) ([]BookingStatusCount, error) {
	var results []BookingStatusCount

	query := r.db.WithContext(ctx).Model(&models.Booking{}).Table("bookings")
	query = r.applyBookingFilter(query, filter)

	if err := query.
		Select("bookings.status, COUNT(*) AS count").
		Group("bookings.status").
		Order("count DESC").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", appErrors.ErrCtxRevenueBookingsByStatus, err)
	}
	return results, nil
}

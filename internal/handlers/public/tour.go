package public

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/repository"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type PublicTourHandler struct {
	service    *services.TourService
	catService *services.CategoryService
}

func NewPublicTourHandler(service *services.TourService, catService *services.CategoryService) *PublicTourHandler {
	return &PublicTourHandler{service: service, catService: catService}
}

func (h *PublicTourHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	minPrice, _ := strconv.ParseFloat(c.Query("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)
	minDuration, _ := strconv.Atoi(c.Query("min_duration"))
	maxDuration, _ := strconv.Atoi(c.Query("max_duration"))

	sortBy, sortOrder := parseSortParam(c.Query("sort"))

	filter := repository.TourFilter{
		Status:       "active",
		CategorySlug: c.Query("category"),
		Search:       c.Query("q"),
		Location:     c.Query("location"),
		MinPrice:     minPrice,
		MaxPrice:     maxPrice,
		MinDuration:  minDuration,
		MaxDuration:  maxDuration,
		SortBy:       sortBy,
		SortOrder:    sortOrder,
		Page:         page,
		Limit:        10,
	}

	tours, total, err := h.service.ListTours(c.Request.Context(), filter)
	if err != nil {
		slog.Error(messages.LogPublicTourListFailed, "error", err)
		c.HTML(http.StatusInternalServerError, "public/pages/error.html", gin.H{
			"title":          messages.ErrInternalServer,
			"message":        messages.ErrInternalServer,
			"user":           middleware.GetCurrentUser(c),
			"nav_categories": middleware.GetNavCategories(c),
		})
		return
	}

	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}

	pages := make([]int, totalPages)
	for i := range pages {
		pages[i] = i + 1
	}

	categories, _ := h.catService.AllFlatCategories(c.Request.Context())

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "public/pages/tours_list.html", gin.H{
		"title":          messages.TitlePublicTours,
		"user":           middleware.GetCurrentUser(c),
		"csrf_token":     middleware.CSRFToken(c),
		"nav_categories": middleware.GetNavCategories(c),
		"flash_success":  flashSuccess,
		"flash_error":    flashError,
		"tours":          tours,
		"total":          total,
		"filter":         filter,
		"categories":     categories,
		"base_url":       buildToursBaseURL(filter),
		"pagination": map[string]any{
			"Page":       page,
			"TotalPages": totalPages,
			"PrevPage":   page - 1,
			"NextPage":   page + 1,
			"Pages":      pages,
		},
	})
}

func (h *PublicTourHandler) Detail(c *gin.Context) {
	slug := c.Param("slug")

	tour, ratingCount, err := h.service.GetPublicTourBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, appErrors.ErrTourNotFound) {
			c.HTML(http.StatusNotFound, "public/pages/error.html", gin.H{
				"title":          "404",
				"status":         404,
				"message":        messages.ErrPublicTourNotFound,
				"user":           middleware.GetCurrentUser(c),
				"nav_categories": middleware.GetNavCategories(c),
			})
			return
		}
		slog.Error(messages.LogPublicTourDetailFailed, "slug", slug, "error", err)
		c.HTML(http.StatusInternalServerError, "public/pages/error.html", gin.H{
			"title":          messages.ErrInternalServer,
			"message":        messages.ErrInternalServer,
			"user":           middleware.GetCurrentUser(c),
			"nav_categories": middleware.GetNavCategories(c),
		})
		return
	}

	var imageURLs []string
	if len(tour.Images) > 0 {
		_ = json.Unmarshal(tour.Images, &imageURLs)
	}

	c.HTML(http.StatusOK, "public/pages/tour_detail.html", gin.H{
		"title":          tour.Title,
		"user":           middleware.GetCurrentUser(c),
		"csrf_token":     middleware.CSRFToken(c),
		"nav_categories": middleware.GetNavCategories(c),
		"tour":           tour,
		"rating_count":   ratingCount,
		"image_urls":     imageURLs,
	})
}

func parseSortParam(sort string) (sortBy, sortOrder string) {
	switch sort {
	case "price_asc":
		return "price", "asc"
	case "price_desc":
		return "price", "desc"
	case "rating":
		return "avg_rating", "desc"
	default:
		return "created_at", "desc"
	}
}

func buildToursBaseURL(filter repository.TourFilter) string {
	v := url.Values{}
	if filter.Search != "" {
		v.Set("q", filter.Search)
	}
	if filter.CategorySlug != "" {
		v.Set("category", filter.CategorySlug)
	}
	if filter.MinPrice > 0 {
		v.Set("min_price", fmt.Sprintf("%.0f", filter.MinPrice))
	}
	if filter.MaxPrice > 0 {
		v.Set("max_price", fmt.Sprintf("%.0f", filter.MaxPrice))
	}
	if filter.Location != "" {
		v.Set("location", filter.Location)
	}
	if filter.MinDuration > 0 {
		v.Set("min_duration", strconv.Itoa(filter.MinDuration))
	}
	if filter.MaxDuration > 0 {
		v.Set("max_duration", strconv.Itoa(filter.MaxDuration))
	}
	if filter.SortBy != "" && filter.SortBy != "created_at" {
		switch filter.SortBy + "_" + filter.SortOrder {
		case "price_asc":
			v.Set("sort", "price_asc")
		case "price_desc":
			v.Set("sort", "price_desc")
		case "avg_rating_desc":
			v.Set("sort", "rating")
		}
	}
	encoded := v.Encode()
	if encoded == "" {
		return "/tours?"
	}
	return "/tours?" + encoded + "&"
}

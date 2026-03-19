package public

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/repository"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	service *services.ReviewService
}

func NewReviewHandler(service *services.ReviewService) *ReviewHandler {
	return &ReviewHandler{service: service}
}

func (h *ReviewHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	filter := repository.ReviewFilter{
		Type:    c.Query("type"),
		Keyword: c.Query("q"),
		Sort:    c.Query("sort"),
		Page:    page,
		Limit:   constants.DefaultPageLimit,
	}

	reviews, total, err := h.service.ListPublicReviews(c.Request.Context(), filter)
	if err != nil {
		slog.Error(messages.LogReviewListFailed, "error", err)
		h.renderError(c, http.StatusInternalServerError, messages.ErrInternalServer)
		return
	}

	totalPages := max(1, (int(total)+constants.DefaultPageLimit-1)/constants.DefaultPageLimit)
	pages := buildPageWindow(page, totalPages)

	c.HTML(http.StatusOK, "public/pages/reviews_list.html", gin.H{
		"title":          messages.TitlePublicReviews,
		"user":           middleware.GetCurrentUser(c),
		"csrf_token":     middleware.CSRFToken(c),
		"nav_categories": middleware.GetNavCategories(c),
		"reviews":        reviews,
		"total":          total,
		"filter":         filter,
		"pagination": map[string]any{
			"Page":       page,
			"TotalPages": totalPages,
			"PrevPage":   max(1, page-1),
			"NextPage":   min(totalPages, page+1),
			"Pages":      pages,
		},
	})
}

func (h *ReviewHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.renderError(c, http.StatusNotFound, messages.ErrReviewNotFound)
		return
	}

	review, comments, err := h.service.GetPublicReview(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, appErrors.ErrReviewNotFound) {
			h.renderError(c, http.StatusNotFound, messages.ErrReviewNotFound)
			return
		}
		slog.Error(messages.LogReviewDetailFailed, "id", id, "error", err)
		h.renderError(c, http.StatusInternalServerError, messages.ErrInternalServer)
		return
	}

	user := middleware.GetCurrentUser(c)
	var userID uint
	if user != nil {
		userID = user.ID
	}
	hasLiked := h.service.HasUserLiked(c.Request.Context(), userID, uint(id))

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "public/pages/review_detail.html", gin.H{
		"title":          review.Title,
		"user":           user,
		"csrf_token":     middleware.CSRFToken(c),
		"nav_categories": middleware.GetNavCategories(c),
		"flash_success":  flashSuccess,
		"flash_error":    flashError,
		"review":         review,
		"comments":       comments,
		"has_liked":      hasLiked,
	})
}

func (h *ReviewHandler) CreateForm(c *gin.Context) {
	flashSuccess, flashError := middleware.GetFlash(c)
	c.HTML(http.StatusOK, "public/pages/review_form.html", gin.H{
		"title":          messages.TitleReviewCreate,
		"user":           middleware.GetCurrentUser(c),
		"csrf_token":     middleware.CSRFToken(c),
		"nav_categories": middleware.GetNavCategories(c),
		"flash_success":  flashSuccess,
		"flash_error":    flashError,
		"is_edit":        false,
	})
}

func (h *ReviewHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	input := services.ReviewCreateInput{
		Title:   c.PostForm("title"),
		Content: c.PostForm("content"),
		Type:    c.PostForm("type"),
		Images:  parseImageURLs(c.PostForm("images")),
	}

	_, err := h.service.CreateReview(c.Request.Context(), user.ID, input)
	if err != nil {
		slog.Error(messages.LogReviewCreateFailed, "user_id", user.ID, "error", err)
		errMsg := messages.ErrReviewCreateFail
		var appErr *appErrors.AppError
		if errors.As(err, &appErr) {
			errMsg = getReviewValidationMsg(input)
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RoutePublicReviewCreate)
		return
	}

	slog.Info("review created", "user_id", user.ID)
	middleware.SetFlashSuccess(c, messages.MsgReviewCreated)
	c.Redirect(http.StatusFound, constants.RouteMyReviews)
}

func (h *ReviewHandler) MyList(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	reviews, total, err := h.service.ListMyReviews(c.Request.Context(), user.ID, page, constants.DefaultPageLimit)
	if err != nil {
		slog.Error(messages.LogReviewMyListFailed, "error", err)
		h.renderError(c, http.StatusInternalServerError, messages.ErrInternalServer)
		return
	}

	totalPages := max(1, (int(total)+constants.DefaultPageLimit-1)/constants.DefaultPageLimit)
	pages := buildPageWindow(page, totalPages)
	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "public/pages/my_reviews_list.html", gin.H{
		"title":          messages.TitleMyReviews,
		"user":           user,
		"csrf_token":     middleware.CSRFToken(c),
		"nav_categories": middleware.GetNavCategories(c),
		"flash_success":  flashSuccess,
		"flash_error":    flashError,
		"reviews":        reviews,
		"total":          total,
		"pagination": map[string]any{
			"Page":       page,
			"TotalPages": totalPages,
			"PrevPage":   max(1, page-1),
			"NextPage":   min(totalPages, page+1),
			"Pages":      pages,
		},
	})
}

func (h *ReviewHandler) EditForm(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.renderError(c, http.StatusNotFound, messages.ErrReviewNotFound)
		return
	}

	review, err := h.service.GetReview(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, appErrors.ErrReviewNotFound) {
			h.renderError(c, http.StatusNotFound, messages.ErrReviewNotFound)
			return
		}
		h.renderError(c, http.StatusInternalServerError, messages.ErrInternalServer)
		return
	}
	if review.UserID != user.ID {
		h.renderError(c, http.StatusForbidden, messages.ErrReviewNotOwner)
		return
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "public/pages/review_form.html", gin.H{
		"title":          messages.TitleReviewEdit,
		"user":           user,
		"csrf_token":     middleware.CSRFToken(c),
		"nav_categories": middleware.GetNavCategories(c),
		"flash_success":  flashSuccess,
		"flash_error":    flashError,
		"review":         review,
		"is_edit":        true,
	})
}

func (h *ReviewHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.renderError(c, http.StatusNotFound, messages.ErrReviewNotFound)
		return
	}

	input := services.ReviewCreateInput{
		Title:   c.PostForm("title"),
		Content: c.PostForm("content"),
		Type:    c.PostForm("type"),
		Images:  parseImageURLs(c.PostForm("images")),
	}

	if err := h.service.UpdateReview(c.Request.Context(), uint(id), user.ID, input); err != nil {
		slog.Error(messages.LogReviewUpdateFailed, "id", id, "user_id", user.ID, "error", err)
		errMsg := messages.ErrReviewUpdateFail
		if errors.Is(err, appErrors.ErrReviewNotFound) {
			errMsg = messages.ErrReviewNotFound
		} else if errors.Is(err, appErrors.ErrReviewNotOwner) {
			errMsg = messages.ErrReviewNotOwner
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d/edit", constants.RouteMyReviews, id))
		return
	}

	slog.Info("review updated", "id", id, "user_id", user.ID)
	middleware.SetFlashSuccess(c, messages.MsgReviewUpdated)
	c.Redirect(http.StatusFound, constants.RouteMyReviews)
}

func (h *ReviewHandler) Delete(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.renderError(c, http.StatusNotFound, messages.ErrReviewNotFound)
		return
	}

	if err := h.service.DeleteReview(c.Request.Context(), uint(id), user.ID); err != nil {
		slog.Error(messages.LogReviewDeleteFailed, "id", id, "user_id", user.ID, "error", err)
		errMsg := messages.ErrReviewDeleteFail
		if errors.Is(err, appErrors.ErrReviewNotFound) {
			errMsg = messages.ErrReviewNotFound
		} else if errors.Is(err, appErrors.ErrReviewNotOwner) {
			errMsg = messages.ErrReviewNotOwner
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RouteMyReviews)
		return
	}

	slog.Info("review deleted", "id", id, "user_id", user.ID)
	middleware.SetFlashSuccess(c, messages.MsgReviewDeleted)
	c.Redirect(http.StatusFound, constants.RouteMyReviews)
}

func (h *ReviewHandler) ToggleLike(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.setFlashAndRedirect(c, messages.ErrReviewNotFound, constants.RoutePublicReviews)
		return
	}

	liked, err := h.service.ToggleLike(c.Request.Context(), user.ID, uint(id))
	if err != nil {
		slog.Error(messages.LogReviewLikeFailed, "review_id", id, "user_id", user.ID, "error", err)
		middleware.SetFlashError(c, messages.ErrLikeFail)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", constants.RoutePublicReviews, id))
		return
	}

	if liked {
		middleware.SetFlashSuccess(c, messages.MsgReviewLiked)
	} else {
		middleware.SetFlashSuccess(c, messages.MsgReviewUnliked)
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", constants.RoutePublicReviews, id))
}

func (h *ReviewHandler) AddComment(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	reviewID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.setFlashAndRedirect(c, messages.ErrReviewNotFound, constants.RoutePublicReviews)
		return
	}

	content := c.PostForm("content")
	if strings.TrimSpace(content) == "" {
		h.setFlashAndRedirect(c, messages.ErrCommentContentReq, fmt.Sprintf("%s/%d", constants.RoutePublicReviews, reviewID))
		return
	}

	if err := h.service.AddComment(c.Request.Context(), user.ID, uint(reviewID), nil, content); err != nil {
		slog.Error(messages.LogReviewCommentFailed, "review_id", reviewID, "user_id", user.ID, "error", err)
		middleware.SetFlashError(c, messages.ErrCommentFail)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", constants.RoutePublicReviews, reviewID))
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgCommentAdded)
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", constants.RoutePublicReviews, reviewID))
}

func (h *ReviewHandler) ReplyComment(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.setFlashAndRedirect(c, messages.ErrCommentNotFound, constants.RoutePublicReviews)
		return
	}

	reviewIDStr := c.PostForm("review_id")
	reviewID, _ := strconv.ParseUint(reviewIDStr, 10, 64)
	if reviewID == 0 {
		h.setFlashAndRedirect(c, messages.ErrReviewNotFound, constants.RoutePublicReviews)
		return
	}

	content := c.PostForm("content")
	if strings.TrimSpace(content) == "" {
		h.setFlashAndRedirect(c, messages.ErrCommentContentReq, fmt.Sprintf("%s/%d", constants.RoutePublicReviews, reviewID))
		return
	}

	parentID := uint(commentID)
	if err := h.service.AddComment(c.Request.Context(), user.ID, uint(reviewID), &parentID, content); err != nil {
		slog.Error(messages.LogReviewCommentFailed, "comment_id", commentID, "user_id", user.ID, "error", err)
		middleware.SetFlashError(c, messages.ErrCommentFail)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", constants.RoutePublicReviews, reviewID))
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgCommentAdded)
	c.Redirect(http.StatusFound, fmt.Sprintf("%s/%d", constants.RoutePublicReviews, reviewID))
}

func (h *ReviewHandler) DeleteComment(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.setFlashAndRedirect(c, messages.ErrCommentNotFound, constants.RoutePublicReviews)
		return
	}

	reviewIDStr := c.PostForm("review_id")
	reviewID, _ := strconv.ParseUint(reviewIDStr, 10, 64)

	isAdmin := user.Role == constants.RoleAdmin
	if err := h.service.DeleteComment(c.Request.Context(), uint(commentID), user.ID, isAdmin); err != nil {
		slog.Error(messages.LogReviewDelCommentFailed, "comment_id", commentID, "user_id", user.ID, "error", err)
		errMsg := messages.ErrCommentDeleteFail
		if errors.Is(err, appErrors.ErrCommentNotFound) {
			errMsg = messages.ErrCommentNotFound
		} else if errors.Is(err, appErrors.ErrCommentNotOwner) {
			errMsg = messages.ErrCommentNotOwner
		}
		redirectURL := constants.RoutePublicReviews
		if reviewID > 0 {
			redirectURL = fmt.Sprintf("%s/%d", constants.RoutePublicReviews, reviewID)
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, redirectURL)
		return
	}

	slog.Info("comment deleted", "comment_id", commentID, "user_id", user.ID)
	redirectURL := constants.RoutePublicReviews
	if reviewID > 0 {
		redirectURL = fmt.Sprintf("%s/%d", constants.RoutePublicReviews, reviewID)
	}
	middleware.SetFlashSuccess(c, messages.MsgCommentDeleted)
	c.Redirect(http.StatusFound, redirectURL)
}

func (h *ReviewHandler) renderError(c *gin.Context, status int, message string) {
	c.HTML(status, "public/pages/error.html", gin.H{
		"title":          message,
		"status":         status,
		"message":        message,
		"user":           middleware.GetCurrentUser(c),
		"nav_categories": middleware.GetNavCategories(c),
	})
}

func (h *ReviewHandler) setFlashAndRedirect(c *gin.Context, msg, url string) {
	middleware.SetFlashError(c, msg)
	c.Redirect(http.StatusFound, url)
}

func parseImageURLs(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, "\n")
	var urls []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			urls = append(urls, p)
		}
	}
	return urls
}

func getReviewValidationMsg(input services.ReviewCreateInput) string {
	if strings.TrimSpace(input.Title) == "" {
		return messages.ErrReviewTitleReq
	}
	if strings.TrimSpace(input.Content) == "" {
		return messages.ErrReviewContentReq
	}
	if len(strings.TrimSpace(input.Content)) < 10 {
		return messages.ErrReviewContentMin
	}
	switch input.Type {
	case constants.ReviewTypePlace, constants.ReviewTypeFood, constants.ReviewTypeNews:
	default:
		return messages.ErrReviewInvalidType
	}
	return messages.ErrReviewCreateFail
}

func buildPageWindow(page, totalPages int) []int {
	const pageWindow = 2
	winStart := page - pageWindow
	if winStart < 1 {
		winStart = 1
	}
	winEnd := page + pageWindow
	if winEnd > totalPages {
		winEnd = totalPages
	}
	pages := make([]int, 0, winEnd-winStart+1)
	for i := winStart; i <= winEnd; i++ {
		pages = append(pages, i)
	}
	return pages
}

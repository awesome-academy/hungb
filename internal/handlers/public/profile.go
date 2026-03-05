package public

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"

	"sun-booking-tours/internal/constants"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileService *services.ProfileService
}

func NewProfileHandler(profileService *services.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

// Show renders GET /profile.
func (h *ProfileHandler) Show(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "public/pages/profile.html", gin.H{
		"title":         messages.TitleProfile,
		"user":          user,
		"csrf_token":    middleware.CSRFToken(c),
		"flash_success": flashSuccess,
		"flash_error":   flashError,
	})
}

// Edit renders GET /profile/edit.
func (h *ProfileHandler) Edit(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	c.HTML(http.StatusOK, "public/pages/profile_edit.html", gin.H{
		"title":      messages.TitleProfileEdit,
		"user":       user,
		"csrf_token": middleware.CSRFToken(c),
		"form": services.ProfileUpdateForm{
			FullName:  user.FullName,
			Phone:     user.Phone,
			AvatarURL: user.AvatarURL,
		},
	})
}

// Update handles POST /profile/edit.
func (h *ProfileHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	var form services.ProfileUpdateForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "public/pages/profile_edit.html", gin.H{
			"title":      messages.TitleProfileEdit,
			"user":       user,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     translateProfileErrors(err),
			"form":       form,
		})
		return
	}

	_, err := h.profileService.UpdateProfile(c.Request.Context(), user.ID, &form)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "public/pages/profile_edit.html", gin.H{
			"title":      messages.TitleProfileEdit,
			"user":       user,
			"csrf_token": middleware.CSRFToken(c),
			"errors":     []string{messages.ErrProfileUpdateFailed},
			"form":       form,
		})
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgProfileUpdateSuccess)
	c.Redirect(http.StatusFound, constants.RouteProfile)
}

// translateProfileErrors converts binding errors to Vietnamese messages.
func translateProfileErrors(err error) []string {
	var valErrs validator.ValidationErrors
	if !errors.As(err, &valErrs) {
		return []string{messages.ErrInvalidForm}
	}

	labels := map[string]string{
		"FullName":  messages.FieldFullName,
		"Phone":     messages.FieldPhone,
		"AvatarURL": messages.FieldAvatarURL,
	}

	msgs := make([]string, 0, len(valErrs))
	for _, fe := range valErrs {
		label := fe.Field()
		if vn, ok := labels[fe.Field()]; ok {
			label = vn
		}
		var msg string
		switch fe.Tag() {
		case "required":
			msg = fmt.Sprintf(messages.ValRequired, label)
		case "min":
			msg = fmt.Sprintf(messages.ValMin, label, fe.Param())
		case "max":
			msg = fmt.Sprintf(messages.ValMax, label, fe.Param())
		default:
			msg = fmt.Sprintf(messages.ValInvalid, label)
		}
		msgs = append(msgs, msg)
	}
	return msgs
}

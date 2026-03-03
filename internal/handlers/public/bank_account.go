package public

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"sun-booking-tours/internal/constants"
	appErrors "sun-booking-tours/internal/errors"
	"sun-booking-tours/internal/messages"
	"sun-booking-tours/internal/middleware"
	"sun-booking-tours/internal/models"
	"sun-booking-tours/internal/services"

	"github.com/gin-gonic/gin"
)

type BankAccountHandler struct {
	service *services.BankAccountService
}

func NewBankAccountHandler(service *services.BankAccountService) *BankAccountHandler {
	return &BankAccountHandler{service: service}
}

// List renders GET /bank-accounts.
func (h *BankAccountHandler) List(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	accounts, err := h.service.ListByUser(c.Request.Context(), user.ID)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrInternalServer)
		c.Redirect(http.StatusFound, constants.RouteProfile)
		return
	}

	flashSuccess, flashError := middleware.GetFlash(c)

	c.HTML(http.StatusOK, "public/pages/bank_accounts_list.html", gin.H{
		"title":         messages.TitleBankAccounts,
		"user":          user,
		"csrf_token":    middleware.CSRFToken(c),
		"accounts":      accounts,
		"flash_success": flashSuccess,
		"flash_error":   flashError,
	})
}

// CreateForm renders GET /bank-accounts/create.
func (h *BankAccountHandler) CreateForm(c *gin.Context) {
	c.HTML(http.StatusOK, "public/pages/bank_account_form.html", gin.H{
		"title":      messages.TitleBankAccountAdd,
		"user":       middleware.GetCurrentUser(c),
		"csrf_token": middleware.CSRFToken(c),
		"is_edit":    false,
	})
}

// Create handles POST /bank-accounts/create.
func (h *BankAccountHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	var form services.BankAccountForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "public/pages/bank_account_form.html", gin.H{
			"title":      messages.TitleBankAccountAdd,
			"user":       user,
			"csrf_token": middleware.CSRFToken(c),
			"is_edit":    false,
			"errors":     translateBankAccountErrors(err),
			"form":       form,
		})
		return
	}

	if err := h.service.Create(c.Request.Context(), user.ID, &form); err != nil {
		c.HTML(http.StatusInternalServerError, "public/pages/bank_account_form.html", gin.H{
			"title":      messages.TitleBankAccountAdd,
			"user":       user,
			"csrf_token": middleware.CSRFToken(c),
			"is_edit":    false,
			"errors":     []string{messages.ErrBankAccountCreateFail},
			"form":       form,
		})
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgBankAccountCreated)
	c.Redirect(http.StatusFound, constants.RouteBankAccounts)
}

// EditForm renders GET /bank-accounts/:id/edit.
func (h *BankAccountHandler) EditForm(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrBankAccountNotFound)
		c.Redirect(http.StatusFound, constants.RouteBankAccounts)
		return
	}

	account, err := h.service.GetByID(c.Request.Context(), uint(id), user.ID)
	if err != nil {
		if appErrors.Is(err, appErrors.ErrForbidden) {
			middleware.SetFlashError(c, messages.ErrBankAccountForbidden)
		} else {
			middleware.SetFlashError(c, messages.ErrBankAccountNotFound)
		}
		c.Redirect(http.StatusFound, constants.RouteBankAccounts)
		return
	}

	c.HTML(http.StatusOK, "public/pages/bank_account_form.html", gin.H{
		"title":      messages.TitleBankAccountEdit,
		"user":       user,
		"csrf_token": middleware.CSRFToken(c),
		"is_edit":    true,
		"account":    account,
		"form": services.BankAccountForm{
			BankName:      account.BankName,
			AccountNumber: account.AccountNumber,
			AccountHolder: account.AccountHolder,
		},
	})
}

// Update handles POST /bank-accounts/:id/edit.
func (h *BankAccountHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrBankAccountNotFound)
		c.Redirect(http.StatusFound, constants.RouteBankAccounts)
		return
	}

	var form services.BankAccountForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusUnprocessableEntity, "public/pages/bank_account_form.html", gin.H{
			"title":      messages.TitleBankAccountEdit,
			"user":       user,
			"csrf_token": middleware.CSRFToken(c),
			"is_edit":    true,
			"account":    &models.BankAccount{ID: uint(id)},
			"errors":     translateBankAccountErrors(err),
			"form":       form,
		})
		return
	}

	if err := h.service.Update(c.Request.Context(), uint(id), user.ID, &form); err != nil {
		errMsg := messages.ErrBankAccountUpdateFail
		if appErrors.Is(err, appErrors.ErrForbidden) {
			errMsg = messages.ErrBankAccountForbidden
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RouteBankAccounts)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgBankAccountUpdated)
	c.Redirect(http.StatusFound, constants.RouteBankAccounts)
}

// Delete handles POST /bank-accounts/:id/delete.
func (h *BankAccountHandler) Delete(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrBankAccountNotFound)
		c.Redirect(http.StatusFound, constants.RouteBankAccounts)
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id), user.ID); err != nil {
		errMsg := messages.ErrBankAccountDeleteFail
		if appErrors.Is(err, appErrors.ErrForbidden) {
			errMsg = messages.ErrBankAccountForbidden
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RouteBankAccounts)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgBankAccountDeleted)
	c.Redirect(http.StatusFound, constants.RouteBankAccounts)
}

// SetDefault handles POST /bank-accounts/:id/set-default.
func (h *BankAccountHandler) SetDefault(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.SetFlashError(c, messages.ErrBankAccountNotFound)
		c.Redirect(http.StatusFound, constants.RouteBankAccounts)
		return
	}

	if err := h.service.SetDefault(c.Request.Context(), uint(id), user.ID); err != nil {
		errMsg := messages.ErrBankAccountUpdateFail
		if appErrors.Is(err, appErrors.ErrForbidden) {
			errMsg = messages.ErrBankAccountForbidden
		}
		middleware.SetFlashError(c, errMsg)
		c.Redirect(http.StatusFound, constants.RouteBankAccounts)
		return
	}

	middleware.SetFlashSuccess(c, messages.MsgBankAccountSetDefault)
	c.Redirect(http.StatusFound, constants.RouteBankAccounts)
}

func translateBankAccountErrors(err error) []string {
	var valErrs validator.ValidationErrors
	if !errors.As(err, &valErrs) {
		return []string{messages.ErrInvalidForm}
	}

	labels := map[string]string{
		"BankName":      messages.FieldBankName,
		"AccountNumber": messages.FieldAccountNumber,
		"AccountHolder": messages.FieldAccountHolder,
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
		case "max":
			msg = fmt.Sprintf(messages.ValMax, label, fe.Param())
		default:
			msg = fmt.Sprintf(messages.ValInvalid, label)
		}
		msgs = append(msgs, msg)
	}
	return msgs
}

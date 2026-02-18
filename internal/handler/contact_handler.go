package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/forever-free1/telegram-go/internal/dto"
	"github.com/forever-free1/telegram-go/internal/service"
)

// ContactHandler 联系人处理器
type ContactHandler struct {
	contactService *service.ContactService
}

// NewContactHandler 创建联系人处理器
func NewContactHandler(contactService *service.ContactService) *ContactHandler {
	return &ContactHandler{contactService: contactService}
}

// SyncContactsRequest 同步通讯录请求
type SyncContactsRequest struct {
	Contacts []service.ContactInput `json:"contacts" binding:"required"`
}

// SyncContacts 同步通讯录
// POST /api/contacts/sync
// @Summary Sync contacts
// @Description Sync contacts from phone book
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body SyncContactsRequest true "Sync contacts request"
// @Success 200 {object} dto.Response
// @Router /api/contacts/sync [post]
func (h *ContactHandler) SyncContacts(c *gin.Context) {
	var req SyncContactsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	currentUser := user.(*service.UserClaims)

	contacts, err := h.contactService.SyncContacts(c.Request.Context(), currentUser.UserID, req.Contacts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(contacts))
}

// GetContacts 获取联系人列表
// GET /api/contacts
// @Summary Get contacts
// @Description Get all contacts for current user
// @Tags contacts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.Response
// @Router /api/contacts [get]
func (h *ContactHandler) GetContacts(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	currentUser := user.(*service.UserClaims)

	contacts, err := h.contactService.GetContacts(c.Request.Context(), currentUser.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Error(500, err.Error()))
		return
	}

	c.JSON(http.StatusOK, dto.Success(contacts))
}

// AddContactRequest 添加联系人请求
type AddContactRequest struct {
	ContactID int64 `json:"contact_id" binding:"required"`
}

// AddContact 添加联系人
// POST /api/contacts
// @Summary Add contact
// @Description Add a user as contact
// @Tags contacts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AddContactRequest true "Add contact request"
// @Success 200 {object} dto.Response
// @Router /api/contacts [post]
func (h *ContactHandler) AddContact(c *gin.Context) {
	var req AddContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	currentUser := user.(*service.UserClaims)

	err := h.contactService.AddContact(c.Request.Context(), currentUser.UserID, req.ContactID)
	if err != nil {
		code := 500
		message := err.Error()
		switch err {
		case service.ErrUserNotFound:
			code = 404
			message = "User not found"
		case service.ErrContactExists:
			code = 409
			message = "Contact already exists"
		}
		c.JSON(code, dto.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}

// DeleteContact 删除联系人
// DELETE /api/contacts/:id
// @Summary Delete contact
// @Description Delete a contact
// @Tags contacts
// @Produce json
// @Security BearerAuth
// @Param id path int true "Contact ID"
// @Success 200 {object} dto.Response
// @Router /api/contacts/{id} [delete]
func (h *ContactHandler) DeleteContact(c *gin.Context) {
	var req struct {
		ContactID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error(400, err.Error()))
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Error(401, "unauthorized"))
		return
	}

	currentUser := user.(*service.UserClaims)

	err := h.contactService.DeleteContact(c.Request.Context(), currentUser.UserID, req.ContactID)
	if err != nil {
		code := 500
		message := err.Error()
		switch err {
		case service.ErrContactNotFound:
			code = 404
			message = "Contact not found"
		}
		c.JSON(code, dto.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, dto.Success(nil))
}

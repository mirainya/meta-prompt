package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"meta-prompt/internal/store"
)

type AdminHandler struct {
	userStore    *store.UserStore
	historyStore *store.HistoryStore
}

func NewAdminHandler(us *store.UserStore, hs *store.HistoryStore) *AdminHandler {
	return &AdminHandler{userStore: us, historyStore: hs}
}

// ========== Dashboard ==========

func (h *AdminHandler) Dashboard(c *gin.Context) {
	userCount, _ := h.userStore.CountAll()
	totalGen, _ := h.historyStore.CountAll()
	todayGen, _ := h.historyStore.CountToday()

	c.JSON(http.StatusOK, gin.H{
		"user_count":        userCount,
		"total_generations": totalGen,
		"today_generations": todayGen,
	})
}

// ========== User Management ==========

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, err := h.userStore.ListAll(100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *AdminHandler) SetUserCredits(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Credits int `json:"credits"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.userStore.SetCredits(id, body.Credits); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "credits updated"})
}

func (h *AdminHandler) SetUserRole(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.userStore.SetRole(id, body.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "role updated"})
}

func (h *AdminHandler) ResetUserPassword(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.userStore.UpdatePassword(id, string(hashed)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password reset"})
}

func (h *AdminHandler) SetUserModels(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var body struct {
		Models []string `json:"models"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.userStore.SetAllowedModels(id, body.Models); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "models updated"})
}

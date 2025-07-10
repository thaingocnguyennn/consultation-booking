// internal/handlers/expert_handler.go
package handlers

import (
	"consultation-booking/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExpertHandler struct {
	expertService *services.ExpertService
}

func NewExpertHandler(expertService *services.ExpertService) *ExpertHandler {
	return &ExpertHandler{
		expertService: expertService,
	}
}

func (h *ExpertHandler) CreateExpert(c *gin.Context) {
	var req struct {
		UserID     uint   `json:"user_id" binding:"required"`
		Speciality string `json:"speciality" binding:"required"`
		Experience int    `json:"experience" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.expertService.CreateExpert(req.UserID, req.Speciality, req.Experience); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Expert created successfully"})
}

func (h *ExpertHandler) GetExperts(c *gin.Context) {
	experts, err := h.expertService.GetExperts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, experts)
}

func (h *ExpertHandler) GetExpertByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expert ID"})
		return
	}

	expert, err := h.expertService.GetExpertByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expert not found"})
		return
	}

	c.JSON(http.StatusOK, expert)
}

func (h *ExpertHandler) CreateAvailableSlot(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Get expert ID from user ID
	var expertID uint
	// This would typically be done through a service method
	// For now, we'll assume userID is the expertID
	expertID = userID.(uint)

	var req services.CreateSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.expertService.CreateAvailableSlot(expertID, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Available slot created successfully"})
}

func (h *ExpertHandler) GetAvailableSlots(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid expert ID"})
		return
	}

	slots, err := h.expertService.GetAvailableSlots(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slots)
}

func (h *ExpertHandler) GetExpertBookings(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Get expert ID from user ID
	var expertID uint
	// This would typically be done through a service method
	expertID = userID.(uint)

	bookings, err := h.expertService.GetExpertBookings(expertID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *ExpertHandler) UpdateBookingStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.expertService.UpdateBookingStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking status updated successfully"})
}

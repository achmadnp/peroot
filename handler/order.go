package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/achmadnp/peroot/model"
	"github.com/achmadnp/peroot/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	whatsAppService service.WhatsAppService
}

func NewOrderHandler(whatsAppService service.WhatsAppService) *OrderHandler {
	return &OrderHandler{whatsAppService: whatsAppService}
}

func (h *OrderHandler) NotifyOrder(c *gin.Context) {
	var req model.OrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("Invalid order request", "error", err, "client_ip", c.ClientIP())
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
			Code:    "VALIDATION_ERROR",
		})
		return
	}

	slog.Info("Received order notification",
		"order_id", req.OrderID,
		"customer", req.CustomerName,
		"total", req.TotalAmount,
		"currency", req.Currency,
		"items_count", len(req.Items),
	)

	messageID, err := h.whatsAppService.SendOrderNotification(c.Request.Context(), &req)
	if err != nil {
		slog.Error("Failed to send WhatsApp notification",
			"order_id", req.OrderID,
			"error", err,
		)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Success: false,
			Error:   "Failed to send WhatsApp notification",
			Code:    "WHATSAPP_SEND_FAILED",
		})
		return
	}

	slog.Info("WhatsApp notification sent successfully",
		"order_id", req.OrderID,
		"whatsapp_message_id", messageID,
	)

	c.JSON(http.StatusOK, model.OrderResponse{
		Success:   true,
		Message:   "Order notification sent successfully via WhatsApp",
		OrderID:   req.OrderID,
		MessageID: messageID,
		Timestamp: time.Now(),
	})
}

func (h *OrderHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "peroot is running",
		"time":    time.Now().Format(time.RFC3339),
	})
}

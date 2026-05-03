package handlers

import (
	"net/http"
	"strings"

	"github.com/YHQZ1/esx/packages/logger"
	matchpb "github.com/YHQZ1/esx/packages/proto/matching"
	riskpb "github.com/YHQZ1/esx/packages/proto/risk"
	"github.com/YHQZ1/esx/services/order-gateway/internal/client"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	risk     *client.RiskClient
	matching *client.MatchingClient
	log      *logger.Logger
}

func NewOrderHandler(risk *client.RiskClient, matching *client.MatchingClient, log *logger.Logger) *OrderHandler {
	return &OrderHandler{risk: risk, matching: matching, log: log}
}

type SubmitOrderRequest struct {
	Symbol      string `json:"symbol" binding:"required"`
	Side        string `json:"side" binding:"required"`
	Type        string `json:"type" binding:"required"`
	TimeInForce string `json:"time_in_force"`
	Quantity    int64  `json:"quantity" binding:"required,min=1"`
	Price       int64  `json:"price"`
}

func (h *OrderHandler) SubmitOrder(c *gin.Context) {
	participantID := c.GetString("participant_id")

	var req SubmitOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	riskSide, matchSide := parseSide(req.Side)
	if riskSide == riskpb.OrderSide_ORDER_SIDE_UNSPECIFIED {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid side, must be BUY or SELL"})
		return
	}

	orderType := parseOrderType(req.Type)

	tif := matchpb.TimeInForce_TIME_IN_FORCE_GTC
	if strings.ToUpper(req.TimeInForce) == "IOC" {
		tif = matchpb.TimeInForce_TIME_IN_FORCE_IOC
	}

	if orderType == matchpb.OrderType_ORDER_TYPE_LIMIT && req.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price required for limit orders"})
		return
	}

	lockID, approved, reason, err := h.risk.CheckAndLock(c.Request.Context(), participantID, req.Symbol, riskSide, req.Quantity, req.Price)
	if err != nil {
		h.log.Error("risk check failed", err, logger.Str("participant_id", participantID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "risk check failed"})
		return
	}

	if !approved {
		c.JSON(http.StatusBadRequest, gin.H{"error": reason})
		return
	}

	orderID, status, err := h.matching.SubmitOrder(c.Request.Context(), participantID, req.Symbol, lockID, matchSide, orderType, tif, req.Quantity, req.Price)
	if err != nil {
		h.log.Error("order submission failed", err, logger.Str("participant_id", participantID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order submission failed"})
		return
	}

	h.log.Info("order submitted",
		logger.Str("participant_id", participantID),
		logger.Str("order_id", orderID),
		logger.Str("symbol", req.Symbol),
		logger.Str("status", status),
	)

	c.JSON(http.StatusCreated, gin.H{
		"order_id": orderID,
		"status":   status,
		"symbol":   req.Symbol,
		"side":     req.Side,
		"quantity": req.Quantity,
		"price":    req.Price,
	})
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	participantID := c.GetString("participant_id")
	orderID := c.Param("id")

	cancelled, reason, err := h.matching.CancelOrder(c.Request.Context(), orderID, participantID)
	if err != nil {
		h.log.Error("cancel order failed", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cancel failed"})
		return
	}

	if !cancelled {
		c.JSON(http.StatusBadRequest, gin.H{"error": reason})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order_id": orderID, "cancelled": true})
}

func parseSide(s string) (riskpb.OrderSide, matchpb.OrderSide) {
	switch strings.ToUpper(s) {
	case "BUY":
		return riskpb.OrderSide_ORDER_SIDE_BUY, matchpb.OrderSide_ORDER_SIDE_BUY
	case "SELL":
		return riskpb.OrderSide_ORDER_SIDE_SELL, matchpb.OrderSide_ORDER_SIDE_SELL
	default:
		return riskpb.OrderSide_ORDER_SIDE_UNSPECIFIED, matchpb.OrderSide_ORDER_SIDE_UNSPECIFIED
	}
}

func parseOrderType(s string) matchpb.OrderType {
	switch strings.ToUpper(s) {
	case "MARKET":
		return matchpb.OrderType_ORDER_TYPE_MARKET
	case "LIMIT":
		return matchpb.OrderType_ORDER_TYPE_LIMIT
	case "STOP":
		return matchpb.OrderType_ORDER_TYPE_STOP
	default:
		return matchpb.OrderType_ORDER_TYPE_LIMIT
	}
}

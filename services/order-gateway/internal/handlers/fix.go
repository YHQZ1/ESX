package handlers

import (
	"net/http"
	"strings"

	"github.com/YHQZ1/esx/packages/logger"
	matchpb "github.com/YHQZ1/esx/packages/proto/matching"
	riskpb "github.com/YHQZ1/esx/packages/proto/risk"
	"github.com/YHQZ1/esx/services/order-gateway/internal/client"
	"github.com/YHQZ1/esx/services/order-gateway/internal/fix"
	"github.com/gin-gonic/gin"
)

type FIXHandler struct {
	registry *client.RegistryClient
	risk     *client.RiskClient
	matching *client.MatchingClient
	log      *logger.Logger
}

func NewFIXHandler(registry *client.RegistryClient, risk *client.RiskClient, matching *client.MatchingClient, log *logger.Logger) *FIXHandler {
	return &FIXHandler{registry: registry, risk: risk, matching: matching, log: log}
}

func (h *FIXHandler) Handle(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusBadRequest, fix.RejectToFIX("", "failed to read body"))
		return
	}

	msg, err := fix.Parse(string(body))
	if err != nil {
		c.String(http.StatusBadRequest, fix.RejectToFIX("", err.Error()))
		return
	}

	clOrdID := msg.GetString(fix.TagClOrdID)

	if msg.MsgType() != fix.MsgTypeNewOrder {
		c.String(http.StatusBadRequest, fix.RejectToFIX(clOrdID, "unsupported message type"))
		return
	}

	apiKey := msg.GetString(fix.TagAPIKey)
	if apiKey == "" {
		apiKey = c.GetHeader("x-api-key")
	}
	if apiKey == "" {
		c.String(http.StatusUnauthorized, fix.RejectToFIX(clOrdID, "missing api key"))
		return
	}

	participantID, isActive, err := h.registry.ValidateAPIKey(c.Request.Context(), apiKey)
	if err != nil || !isActive || participantID == "" {
		c.String(http.StatusUnauthorized, fix.RejectToFIX(clOrdID, "invalid api key"))
		return
	}

	symbol := msg.GetString(fix.TagSymbol)
	if symbol == "" {
		c.String(http.StatusBadRequest, fix.RejectToFIX(clOrdID, "missing symbol"))
		return
	}

	fixSide := msg.GetString(fix.TagSide)
	var side riskpb.OrderSide
	var matchSide matchpb.OrderSide
	switch fixSide {
	case fix.SideBuy:
		side = riskpb.OrderSide_ORDER_SIDE_BUY
		matchSide = matchpb.OrderSide_ORDER_SIDE_BUY
	case fix.SideSell:
		side = riskpb.OrderSide_ORDER_SIDE_SELL
		matchSide = matchpb.OrderSide_ORDER_SIDE_SELL
	default:
		c.String(http.StatusBadRequest, fix.RejectToFIX(clOrdID, "invalid side"))
		return
	}

	quantity, err := msg.GetInt(fix.TagOrderQty)
	if err != nil || quantity <= 0 {
		c.String(http.StatusBadRequest, fix.RejectToFIX(clOrdID, "invalid quantity"))
		return
	}

	price, _ := msg.GetInt(fix.TagPrice)

	fixOrdType := msg.GetString(fix.TagOrdType)
	var orderType matchpb.OrderType
	switch fixOrdType {
	case fix.OrdTypeMarket:
		orderType = matchpb.OrderType_ORDER_TYPE_MARKET
	case fix.OrdTypeLimit:
		orderType = matchpb.OrderType_ORDER_TYPE_LIMIT
	case fix.OrdTypeStop:
		orderType = matchpb.OrderType_ORDER_TYPE_STOP
	default:
		orderType = matchpb.OrderType_ORDER_TYPE_LIMIT
	}

	tif := matchpb.TimeInForce_TIME_IN_FORCE_GTC
	if msg.GetString(fix.TagTimeInForce) == fix.TimeInForceIOC {
		tif = matchpb.TimeInForce_TIME_IN_FORCE_IOC
	}

	lockID, approved, reason, err := h.risk.CheckAndLock(c.Request.Context(), participantID, symbol, side, quantity, price)
	if err != nil {
		h.log.Error("risk check failed", err)
		c.String(http.StatusInternalServerError, fix.RejectToFIX(clOrdID, "risk check failed"))
		return
	}
	if !approved {
		c.String(http.StatusBadRequest, fix.RejectToFIX(clOrdID, reason))
		return
	}

	orderID, status, err := h.matching.SubmitOrder(c.Request.Context(), participantID, symbol, lockID, matchSide, orderType, tif, quantity, price)
	if err != nil {
		h.log.Error("order submission failed", err)
		c.String(http.StatusInternalServerError, fix.RejectToFIX(clOrdID, "order submission failed"))
		return
	}

	fixSideStr := fix.SideBuy
	if strings.ToUpper(fixSide) == fix.SideSell {
		fixSideStr = fix.SideSell
	}

	c.String(http.StatusOK, fix.NewOrderToFIX(orderID, symbol, fixSideStr, status, clOrdID))
}

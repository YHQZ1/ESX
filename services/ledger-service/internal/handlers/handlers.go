package handlers

import (
	"net/http"
	"strconv"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/ledger-service/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	db  *db.Queries
	log *logger.Logger
}

func New(database *db.Queries, log *logger.Logger) *Handler {
	return &Handler{db: database, log: log}
}

func (h *Handler) GetBalance(c *gin.Context) {
	participantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid participant_id"})
		return
	}

	balance, err := h.db.GetCashBalance(c.Request.Context(), participantID)
	if err != nil {
		h.log.Error("failed to get balance", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participant_id": participantID,
		"balance":        balance,
		"currency":       "INR",
	})
}

func (h *Handler) GetPositions(c *gin.Context) {
	participantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid participant_id"})
		return
	}

	positions, err := h.db.GetSecuritiesPositions(c.Request.Context(), participantID)
	if err != nil {
		h.log.Error("failed to get positions", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get positions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participant_id": participantID,
		"positions":      positions,
	})
}

func (h *Handler) GetCashTransactions(c *gin.Context) {
	participantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid participant_id"})
		return
	}

	limit := int64(50)
	offset := int64(0)

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.ParseInt(l, 10, 64); err == nil {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.ParseInt(o, 10, 64); err == nil {
			offset = v
		}
	}

	entries, err := h.db.GetCashEntries(c.Request.Context(), participantID, limit, offset)
	if err != nil {
		h.log.Error("failed to get cash entries", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participant_id": participantID,
		"entries":        entries,
	})
}

func (h *Handler) GetSecuritiesTransactions(c *gin.Context) {
	participantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid participant_id"})
		return
	}

	limit := int64(50)
	offset := int64(0)

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.ParseInt(l, 10, 64); err == nil {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.ParseInt(o, 10, 64); err == nil {
			offset = v
		}
	}

	entries, err := h.db.GetSecuritiesEntries(c.Request.Context(), participantID, limit, offset)
	if err != nil {
		h.log.Error("failed to get securities entries", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participant_id": participantID,
		"entries":        entries,
	})
}

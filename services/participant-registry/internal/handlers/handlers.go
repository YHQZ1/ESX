package handlers

import (
	"database/sql"
	"net/http"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/participant-registry/internal/db"
	"github.com/YHQZ1/esx/services/participant-registry/internal/lib"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	db  db.Querier
	log *logger.Logger
}

func New(database db.Querier, log *logger.Logger) *Handler {
	return &Handler{db: database, log: log}
}

func (h *Handler) Register(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	participant, err := h.db.CreateParticipant(c.Request.Context(), req.Name, req.Email)
	if err != nil {
		h.log.Error("failed to create participant", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create participant"})
		return
	}

	_, err = h.db.CreateCashAccount(c.Request.Context(), participant.ID)
	if err != nil {
		h.log.Error("failed to create cash account", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create cash account"})
		return
	}

	rawKey, err := lib.GenerateAPIKey()
	if err != nil {
		h.log.Error("failed to generate api key", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate api key"})
		return
	}

	_, err = h.db.CreateAPIKey(c.Request.Context(), participant.ID, lib.HashAPIKey(rawKey))
	if err != nil {
		h.log.Error("failed to create api key", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create api key"})
		return
	}

	h.log.Info("participant registered",
		logger.Str("participant_id", participant.ID.String()),
		logger.Str("email", participant.Email),
	)

	c.JSON(http.StatusCreated, gin.H{
		"participant_id": participant.ID,
		"name":           participant.Name,
		"email":          participant.Email,
		"api_key":        rawKey,
	})
}

func (h *Handler) Deposit(c *gin.Context) {
	participantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid participant id"})
		return
	}

	var req struct {
		Amount int64 `json:"amount" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.db.Deposit(c.Request.Context(), participantID, req.Amount)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "participant not found"})
			return
		}
		h.log.Error("failed to deposit", err, logger.Str("participant_id", participantID.String()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deposit"})
		return
	}

	h.log.Info("deposit successful",
		logger.Str("participant_id", participantID.String()),
		logger.Int64("amount", req.Amount),
		logger.Int64("balance", account.Balance),
	)

	c.JSON(http.StatusOK, gin.H{
		"participant_id": account.ParticipantID,
		"balance":        account.Balance,
		"currency":       account.Currency,
	})
}

func (h *Handler) GetAccount(c *gin.Context) {
	participantID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid participant id"})
		return
	}

	participant, err := h.db.GetParticipantByID(c.Request.Context(), participantID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "participant not found"})
			return
		}
		h.log.Error("failed to get participant", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get participant"})
		return
	}

	cashAccount, err := h.db.GetCashAccount(c.Request.Context(), participantID)
	if err != nil {
		h.log.Error("failed to get cash account", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get cash account"})
		return
	}

	securities, err := h.db.GetAllSecuritiesAccounts(c.Request.Context(), participantID)
	if err != nil {
		h.log.Error("failed to get securities accounts", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get securities accounts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participant": participant,
		"cash":        cashAccount,
		"securities":  securities,
	})
}

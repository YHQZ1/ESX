package journal

import (
	"context"
	"fmt"

	"github.com/YHQZ1/esx/packages/logger"
	"github.com/YHQZ1/esx/services/ledger-service/internal/db"
	"github.com/google/uuid"
)

type Journal struct {
	db  db.Querier
	log *logger.Logger
}

func New(database db.Querier, log *logger.Logger) *Journal {
	return &Journal{db: database, log: log}
}

func (j *Journal) Record(ctx context.Context, arg RecordParams) error {
	totalCash := arg.Price * arg.Quantity

	buyerCashAfter, err := j.db.GetCashBalance(ctx, arg.BuyerID)
	if err != nil {
		return fmt.Errorf("failed to get buyer cash balance: %w", err)
	}

	sellerCashAfter, err := j.db.GetCashBalance(ctx, arg.SellerID)
	if err != nil {
		return fmt.Errorf("failed to get seller cash balance: %w", err)
	}

	if _, err := j.db.CreateCashEntry(ctx, db.CreateCashEntryParams{
		TradeID:       arg.TradeID,
		ParticipantID: arg.BuyerID,
		EntryType:     "debit",
		Amount:        totalCash,
		BalanceAfter:  buyerCashAfter,
	}); err != nil {
		return fmt.Errorf("failed to write buyer cash debit: %w", err)
	}

	if _, err := j.db.CreateCashEntry(ctx, db.CreateCashEntryParams{
		TradeID:       arg.TradeID,
		ParticipantID: arg.SellerID,
		EntryType:     "credit",
		Amount:        totalCash,
		BalanceAfter:  sellerCashAfter,
	}); err != nil {
		return fmt.Errorf("failed to write seller cash credit: %w", err)
	}

	positions, err := j.db.GetSecuritiesPositions(ctx, arg.SellerID)
	if err != nil {
		return fmt.Errorf("failed to get seller positions: %w", err)
	}
	sellerSharesAfter := positions[arg.Symbol]

	buyerPositions, err := j.db.GetSecuritiesPositions(ctx, arg.BuyerID)
	if err != nil {
		return fmt.Errorf("failed to get buyer positions: %w", err)
	}
	buyerSharesAfter := buyerPositions[arg.Symbol]

	if _, err := j.db.CreateSecuritiesEntry(ctx, db.CreateSecuritiesEntryParams{
		TradeID:       arg.TradeID,
		ParticipantID: arg.SellerID,
		Symbol:        arg.Symbol,
		EntryType:     "debit",
		Quantity:      arg.Quantity,
		BalanceAfter:  sellerSharesAfter,
	}); err != nil {
		return fmt.Errorf("failed to write seller shares debit: %w", err)
	}

	if _, err := j.db.CreateSecuritiesEntry(ctx, db.CreateSecuritiesEntryParams{
		TradeID:       arg.TradeID,
		ParticipantID: arg.BuyerID,
		Symbol:        arg.Symbol,
		EntryType:     "credit",
		Quantity:      arg.Quantity,
		BalanceAfter:  buyerSharesAfter,
	}); err != nil {
		return fmt.Errorf("failed to write buyer shares credit: %w", err)
	}

	j.log.Info("journal entries recorded",
		logger.Str("trade_id", arg.TradeID.String()),
		logger.Str("symbol", arg.Symbol),
		logger.Int64("price", arg.Price),
		logger.Int64("quantity", arg.Quantity),
		logger.Int64("total_cash", totalCash),
	)

	return nil
}

type RecordParams struct {
	TradeID  uuid.UUID
	Symbol   string
	BuyerID  uuid.UUID
	SellerID uuid.UUID
	Price    int64
	Quantity int64
}

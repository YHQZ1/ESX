package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Querier interface {
	CreateCashEntry(ctx context.Context, arg CreateCashEntryParams) (CashEntry, error)
	CreateSecuritiesEntry(ctx context.Context, arg CreateSecuritiesEntryParams) (SecuritiesEntry, error)
	GetCashEntries(ctx context.Context, participantID uuid.UUID, limit, offset int64) ([]CashEntry, error)
	GetSecuritiesEntries(ctx context.Context, participantID uuid.UUID, limit, offset int64) ([]SecuritiesEntry, error)
	GetCashBalance(ctx context.Context, participantID uuid.UUID) (int64, error)
	GetSecuritiesPositions(ctx context.Context, participantID uuid.UUID) (map[string]int64, error)
}

type CreateCashEntryParams struct {
	TradeID       uuid.UUID
	ParticipantID uuid.UUID
	EntryType     string
	Amount        int64
	BalanceAfter  int64
}

type CreateSecuritiesEntryParams struct {
	TradeID       uuid.UUID
	ParticipantID uuid.UUID
	Symbol        string
	EntryType     string
	Quantity      int64
	BalanceAfter  int64
}

type Queries struct {
	ledgerDB      *sql.DB
	participantDB *sql.DB
}

func New(ledgerDB, participantDB *sql.DB) *Queries {
	return &Queries{ledgerDB: ledgerDB, participantDB: participantDB}
}

func (q *Queries) CreateCashEntry(ctx context.Context, arg CreateCashEntryParams) (CashEntry, error) {
	var e CashEntry
	err := q.ledgerDB.QueryRowContext(ctx,
		`INSERT INTO cash_journal (trade_id, participant_id, entry_type, amount, balance_after)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, trade_id, participant_id, entry_type, amount, balance_after, created_at`,
		arg.TradeID, arg.ParticipantID, arg.EntryType, arg.Amount, arg.BalanceAfter,
	).Scan(&e.ID, &e.TradeID, &e.ParticipantID, &e.EntryType, &e.Amount, &e.BalanceAfter, &e.CreatedAt)
	return e, err
}

func (q *Queries) CreateSecuritiesEntry(ctx context.Context, arg CreateSecuritiesEntryParams) (SecuritiesEntry, error) {
	var e SecuritiesEntry
	err := q.ledgerDB.QueryRowContext(ctx,
		`INSERT INTO securities_journal (trade_id, participant_id, symbol, entry_type, quantity, balance_after)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, trade_id, participant_id, symbol, entry_type, quantity, balance_after, created_at`,
		arg.TradeID, arg.ParticipantID, arg.Symbol, arg.EntryType, arg.Quantity, arg.BalanceAfter,
	).Scan(&e.ID, &e.TradeID, &e.ParticipantID, &e.Symbol, &e.EntryType, &e.Quantity, &e.BalanceAfter, &e.CreatedAt)
	return e, err
}

func (q *Queries) GetCashEntries(ctx context.Context, participantID uuid.UUID, limit, offset int64) ([]CashEntry, error) {
	rows, err := q.ledgerDB.QueryContext(ctx,
		`SELECT id, trade_id, participant_id, entry_type, amount, balance_after, created_at
		 FROM cash_journal WHERE participant_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		participantID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []CashEntry
	for rows.Next() {
		var e CashEntry
		if err := rows.Scan(&e.ID, &e.TradeID, &e.ParticipantID, &e.EntryType, &e.Amount, &e.BalanceAfter, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (q *Queries) GetSecuritiesEntries(ctx context.Context, participantID uuid.UUID, limit, offset int64) ([]SecuritiesEntry, error) {
	rows, err := q.ledgerDB.QueryContext(ctx,
		`SELECT id, trade_id, participant_id, symbol, entry_type, quantity, balance_after, created_at
		 FROM securities_journal WHERE participant_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		participantID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []SecuritiesEntry
	for rows.Next() {
		var e SecuritiesEntry
		if err := rows.Scan(&e.ID, &e.TradeID, &e.ParticipantID, &e.Symbol, &e.EntryType, &e.Quantity, &e.BalanceAfter, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (q *Queries) GetCashBalance(ctx context.Context, participantID uuid.UUID) (int64, error) {
	var balance int64
	err := q.participantDB.QueryRowContext(ctx,
		`SELECT balance FROM cash_accounts WHERE participant_id = $1`,
		participantID,
	).Scan(&balance)
	return balance, err
}

func (q *Queries) GetSecuritiesPositions(ctx context.Context, participantID uuid.UUID) (map[string]int64, error) {
	rows, err := q.participantDB.QueryContext(ctx,
		`SELECT symbol, quantity FROM securities_accounts WHERE participant_id = $1`,
		participantID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	positions := make(map[string]int64)
	for rows.Next() {
		var symbol string
		var quantity int64
		if err := rows.Scan(&symbol, &quantity); err != nil {
			return nil, err
		}
		positions[symbol] = quantity
	}
	return positions, rows.Err()
}

package db

import (
	"context"
	"database/sql"
	"errors"
)

var ErrNotFound = errors.New("not found")

func (d *DB) GetCashAccount(ctx context.Context, participantID string) (*CashAccount, error) {
	row := d.conn.QueryRowContext(ctx, `
		SELECT id, participant_id, balance, locked, currency, created_at, updated_at
		FROM cash_accounts
		WHERE participant_id = $1
	`, participantID)

	a := &CashAccount{}
	err := row.Scan(&a.ID, &a.ParticipantID, &a.Balance, &a.Locked, &a.Currency, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (d *DB) GetSecuritiesAccount(ctx context.Context, participantID, symbol string) (*SecuritiesAccount, error) {
	row := d.conn.QueryRowContext(ctx, `
		SELECT id, participant_id, symbol, quantity, locked, created_at, updated_at
		FROM securities_accounts
		WHERE participant_id = $1 AND symbol = $2
	`, participantID, symbol)

	a := &SecuritiesAccount{}
	err := row.Scan(&a.ID, &a.ParticipantID, &a.Symbol, &a.Quantity, &a.Locked, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (d *DB) CreateLockAndIncrementCashLocked(ctx context.Context, lockID, participantID, symbol string, quantity, price, lockedAmount int64) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO locks (id, participant_id, symbol, side, quantity, price, locked_amount, status)
		VALUES ($1, $2, $3, 'BUY', $4, $5, $6, 'active')
	`, lockID, participantID, symbol, quantity, price, lockedAmount)
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE cash_accounts
		SET locked = locked + $1, updated_at = NOW()
		WHERE participant_id = $2 AND (balance - locked) >= $1
	`, lockedAmount, participantID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("insufficient available cash balance")
	}

	return tx.Commit()
}

func (d *DB) CreateLockAndIncrementSharesLocked(ctx context.Context, lockID, participantID, symbol string, quantity, price int64) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO locks (id, participant_id, symbol, side, quantity, price, locked_amount, status)
		VALUES ($1, $2, $3, 'SELL', $4, $5, $4, 'active')
	`, lockID, participantID, symbol, quantity, price)
	if err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE securities_accounts
		SET locked = locked + $1, updated_at = NOW()
		WHERE participant_id = $2 AND symbol = $3 AND (quantity - locked) >= $1
	`, quantity, participantID, symbol)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("insufficient available share position")
	}

	return tx.Commit()
}

func (d *DB) GetLock(ctx context.Context, lockID string) (*Lock, error) {
	row := d.conn.QueryRowContext(ctx, `
		SELECT id, participant_id, symbol, side, quantity, price, locked_amount, status, created_at, updated_at
		FROM locks
		WHERE id = $1
	`, lockID)

	l := &Lock{}
	err := row.Scan(&l.ID, &l.ParticipantID, &l.Symbol, &l.Side, &l.Quantity, &l.Price, &l.LockedAmount, &l.Status, &l.CreatedAt, &l.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return l, err
}

func (d *DB) ReleaseCashLock(ctx context.Context, lockID string, releaseAmount int64) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lock Lock
	err = tx.QueryRowContext(ctx, `
		SELECT participant_id, locked_amount, status FROM locks WHERE id = $1 FOR UPDATE
	`, lockID).Scan(&lock.ParticipantID, &lock.LockedAmount, &lock.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if lock.Status != LockStatusActive {
		return errors.New("lock is not active")
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE locks SET status = 'released', updated_at = NOW() WHERE id = $1
	`, lockID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE cash_accounts SET locked = locked - $1, updated_at = NOW()
		WHERE participant_id = $2
	`, releaseAmount, lock.ParticipantID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) ReleaseSharesLock(ctx context.Context, lockID string, releaseQuantity int64) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lock Lock
	err = tx.QueryRowContext(ctx, `
		SELECT participant_id, symbol, locked_amount, status FROM locks WHERE id = $1 FOR UPDATE
	`, lockID).Scan(&lock.ParticipantID, &lock.Symbol, &lock.LockedAmount, &lock.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if lock.Status != LockStatusActive {
		return errors.New("lock is not active")
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE locks SET status = 'released', updated_at = NOW() WHERE id = $1
	`, lockID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE securities_accounts SET locked = locked - $1, updated_at = NOW()
		WHERE participant_id = $2 AND symbol = $3
	`, releaseQuantity, lock.ParticipantID, lock.Symbol)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) ConsumeLock(ctx context.Context, lockID string) error {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lock Lock
	err = tx.QueryRowContext(ctx, `
		SELECT participant_id, symbol, side, locked_amount, status FROM locks WHERE id = $1 FOR UPDATE
	`, lockID).Scan(&lock.ParticipantID, &lock.Symbol, &lock.Side, &lock.LockedAmount, &lock.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if lock.Status != LockStatusActive {
		return errors.New("lock is not active")
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE locks SET status = 'consumed', updated_at = NOW() WHERE id = $1
	`, lockID)
	if err != nil {
		return err
	}

	if lock.Side == "BUY" {
		_, err = tx.ExecContext(ctx, `
			UPDATE cash_accounts
			SET balance = balance - $1, locked = locked - $1, updated_at = NOW()
			WHERE participant_id = $2
		`, lock.LockedAmount, lock.ParticipantID)
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE securities_accounts
			SET quantity = quantity - $1, locked = locked - $1, updated_at = NOW()
			WHERE participant_id = $2 AND symbol = $3
		`, lock.LockedAmount, lock.ParticipantID, lock.Symbol)
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

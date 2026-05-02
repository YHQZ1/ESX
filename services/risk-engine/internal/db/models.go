package db

import (
	"database/sql"
	"time"
)

type LockStatus string

const (
	LockStatusActive   LockStatus = "active"
	LockStatusReleased LockStatus = "released"
	LockStatusConsumed LockStatus = "consumed"
)

type Lock struct {
	ID            string
	ParticipantID string
	Symbol        string
	Side          string
	Quantity      int64
	Price         int64
	LockedAmount  int64
	Status        LockStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CashAccount struct {
	ID            string
	ParticipantID string
	Balance       int64
	Locked        int64
	Currency      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SecuritiesAccount struct {
	ID            string
	ParticipantID string
	Symbol        string
	Quantity      int64
	Locked        int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type DB struct {
	conn *sql.DB
}

func New(conn *sql.DB) *DB {
	return &DB{conn: conn}
}

func (d *DB) Conn() *sql.DB {
	return d.conn
}
